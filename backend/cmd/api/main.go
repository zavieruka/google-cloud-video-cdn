package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/zavieruka/video-platform/backend/internal/config"
	"github.com/zavieruka/video-platform/backend/internal/database"
	"github.com/zavieruka/video-platform/backend/internal/handlers"
	"github.com/zavieruka/video-platform/backend/internal/hls"
	"github.com/zavieruka/video-platform/backend/internal/middleware"
	"github.com/zavieruka/video-platform/backend/internal/pubsub"
	"github.com/zavieruka/video-platform/backend/internal/services"
	"github.com/zavieruka/video-platform/backend/internal/storage"
	"github.com/zavieruka/video-platform/backend/internal/validation"
	"github.com/zavieruka/video-platform/backend/internal/version"
)

func main() {
	// Load .env file if it exists (for local development)
	// In Cloud Run, environment variables are set directly
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Starting Video Platform API (Environment: %s)", cfg.Environment)

	// Initialize GCP clients
	if err := cfg.InitializeGCPClients(ctx); err != nil {
		log.Fatalf("Failed to initialize GCP clients: %v", err)
	}
	defer func() {
		if err := cfg.Close(); err != nil {
			log.Printf("Error closing GCP clients: %v", err)
		}
	}()

	log.Println("GCP clients initialized successfully")

	// Initialize Pub/Sub publisher. Hold it as the services.Publisher interface so
	// that when auto-processing is disabled or init fails we pass a genuine nil
	// interface — a typed nil *pubsub.Publisher would be non-nil as an interface
	// and panic when the service tried to publish.
	var publisher services.Publisher
	autoProcessing := cfg.EnableAutoProcessing
	if cfg.EnableAutoProcessing {
		p, err := pubsub.NewPublisher(ctx, cfg.GCPProjectID, map[string]string{
			"video-uploaded":      cfg.PubSubVideoUploadedTopic,
			"processing-complete": cfg.PubSubProcessingCompleteTopic,
		})
		if err != nil {
			log.Printf("Warning: Failed to initialize Pub/Sub publisher: %v", err)
			log.Println("Auto-processing will be disabled. Videos can still be uploaded.")
			autoProcessing = false
		} else {
			log.Println("Pub/Sub publisher initialized successfully")
			publisher = p
			defer p.Close()
		}
	} else {
		log.Println("Auto-processing is disabled")
	}

	// Initialize services
	videoStorage := storage.NewGCSVideoStorage(cfg.StorageClient, cfg.SourceBucketName, cfg.ServiceAccountEmail)
	processedStorage := storage.NewGCSVideoStorage(cfg.StorageClient, cfg.ProcessedBucketName, cfg.ServiceAccountEmail)
	videoRepository := database.NewFirestoreVideoRepository(cfg.FirestoreClient)
	videoValidator := validation.NewVideoValidator(cfg.MaxUploadSizeMB, cfg.AllowedVideoFormats)
	videoService := services.NewVideoService(
		videoRepository,
		videoStorage,
		videoValidator,
		cfg.UploadURLExpiryHrs,
		publisher,
		cfg.SourceBucketName,
		autoProcessing,
	)

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(cfg)
	videoHandler := handlers.NewVideoHandler(videoService)
	hlsHandler := handlers.NewHLSHandler(
		videoService,
		processedStorage,
		hls.MasterPlaylistName,
		time.Duration(cfg.HLSSignedURLExpiryHrs)*time.Hour,
	)

	log.Println("Services and handlers initialized successfully")

	// Set up HTTP routes
	mux := http.NewServeMux()

	// Health check endpoints (for Cloud Run)
	mux.HandleFunc("/health", healthHandler.HandleHealth)
	mux.HandleFunc("/ready", healthHandler.HandleReady)

	// Video endpoints
	mux.HandleFunc("POST /api/v1/videos/upload-url", videoHandler.RequestUploadURL)
	mux.HandleFunc("POST /api/v1/videos/{id}/confirm", videoHandler.ConfirmUpload)
	mux.HandleFunc("POST /api/v1/videos/{id}/fail", videoHandler.FailUpload)
	mux.HandleFunc("GET /api/v1/videos/{id}", videoHandler.GetVideo)
	mux.HandleFunc("GET /api/v1/videos", videoHandler.ListVideos)
	mux.HandleFunc("DELETE /api/v1/videos/{id}", videoHandler.DeleteVideo)

	// HLS delivery: playlists are served from the private processed bucket;
	// segments are fetched directly from GCS via the signed URLs embedded in the
	// rendition playlists.
	mux.HandleFunc("GET /api/v1/videos/{id}/hls/{file}", hlsHandler.ServePlaylist)

	// Root endpoint
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		fmt.Fprintf(w, "Video Platform API - %s\n", version.Version)
	})

	cors := middleware.CORSMiddleware(cfg.CORSAllowedOrigins)
	handler := middleware.LoggingMiddleware(middleware.RecoveryMiddleware(cors(mux)))
	server := &http.Server{
		Addr:         cfg.GetAddress(),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Server starting on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Server shutting down...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}
