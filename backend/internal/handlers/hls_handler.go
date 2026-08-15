package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/zavieruka/video-platform/backend/internal/errors"
	"github.com/zavieruka/video-platform/backend/internal/hls"
	"github.com/zavieruka/video-platform/backend/internal/models"
)

// HLSStorage is the narrow slice of the processed-bucket storage the HLS
// delivery layer needs: read the small playlists, and mint signed URLs for the
// segments they reference. Keeping it minimal lets the handler be tested
// without the full GCS storage.
type HLSStorage interface {
	ReadFile(ctx context.Context, objectName string) ([]byte, error)
	GenerateSignedDownloadURL(ctx context.Context, objectName string, expiry time.Duration) (string, error)
}

// VideoLookup resolves a video so the delivery layer can authorize a request
// (video exists and is ready) before serving any playlist.
type VideoLookup interface {
	GetVideo(ctx context.Context, videoID string) (*models.Video, error)
}

// HLSHandler serves HLS playlists for ready videos out of the private processed
// bucket. The master playlist is passed through unchanged (its relative
// rendition references resolve back to this same endpoint); rendition playlists
// have their segment/init references rewritten to short-lived signed GCS URLs
// so the player fetches segments directly from storage without the bucket ever
// being public.
type HLSHandler struct {
	videos       VideoLookup
	storage      HLSStorage
	manifestName string
	segmentTTL   time.Duration
}

func NewHLSHandler(videos VideoLookup, storage HLSStorage, manifestName string, segmentTTL time.Duration) *HLSHandler {
	return &HLSHandler{
		videos:       videos,
		storage:      storage,
		manifestName: manifestName,
		segmentTTL:   segmentTTL,
	}
}

// ServePlaylist handles GET /api/v1/videos/{id}/hls/{file}. Only .m3u8
// playlists are served here; segments are fetched straight from GCS via the
// signed URLs embedded in the rendition playlists.
func (h *HLSHandler) ServePlaylist(w http.ResponseWriter, r *http.Request) {
	videoID := r.PathValue("id")
	file := r.PathValue("file")
	if videoID == "" || file == "" {
		h.respondError(w, errors.NewBadRequestError("Video ID and playlist name are required"))
		return
	}

	if !strings.HasSuffix(file, ".m3u8") {
		h.respondError(w, errors.NewBadRequestError("Only .m3u8 playlists are served here"))
		return
	}
	// The playlist name is a single object under the video's prefix; reject
	// anything that tries to escape it.
	if strings.Contains(file, "/") || strings.Contains(file, "..") {
		h.respondError(w, errors.NewBadRequestError("Invalid playlist name"))
		return
	}

	video, err := h.videos.GetVideo(r.Context(), videoID)
	if err != nil {
		h.respondError(w, err)
		return
	}
	if video.Status != models.StatusReady {
		h.respondError(w, errors.NewConflictError("Video is not ready for playback"))
		return
	}

	objectName := videoID + "/" + file
	data, err := h.storage.ReadFile(r.Context(), objectName)
	if err != nil {
		h.respondError(w, err)
		return
	}

	// Master playlist: serve as-is. Its relative rendition references resolve
	// against this URL back to the rendition endpoint below.
	if file == h.manifestName {
		h.writePlaylist(w, data)
		return
	}

	// Rendition playlist: sign every segment/init reference.
	rewritten, err := hls.RewritePlaylist(string(data), h.segmentResolver(r.Context(), videoID))
	if err != nil {
		h.respondError(w, errors.NewInternalError("Failed to prepare playlist", err))
		return
	}

	h.writePlaylist(w, []byte(rewritten))
}

// segmentResolver signs a segment reference relative to the video's prefix,
// leaving any already-absolute URL untouched. Transcoder playlists can refer
// to one fMP4 object many times through byte ranges, so reuse its signed URL
// for this response instead of making the same IAM signing call repeatedly.
func (h *HLSHandler) segmentResolver(ctx context.Context, videoID string) hls.URIResolver {
	signedURLs := make(map[string]string)

	return func(uri string) (string, error) {
		if strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://") {
			return uri, nil
		}
		if signedURL, ok := signedURLs[uri]; ok {
			return signedURL, nil
		}

		signedURL, err := h.storage.GenerateSignedDownloadURL(ctx, videoID+"/"+uri, h.segmentTTL)
		if err != nil {
			return "", err
		}
		signedURLs[uri] = signedURL

		return signedURL, nil
	}
}

func (h *HLSHandler) writePlaylist(w http.ResponseWriter, body []byte) {
	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	// Rendition playlists embed signed URLs that expire; keep the cache window
	// well under the signed-URL lifetime.
	w.Header().Set("Cache-Control", "private, max-age=60")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}

func (h *HLSHandler) respondError(w http.ResponseWriter, err error) {
	appErr, ok := err.(*errors.AppError)
	if !ok {
		appErr = errors.NewInternalError("An unexpected error occurred", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.StatusCode)

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	if encErr := encoder.Encode(appErr); encErr != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
