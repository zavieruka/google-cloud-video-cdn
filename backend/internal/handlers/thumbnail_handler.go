package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/zavieruka/video-platform/backend/internal/errors"
	"github.com/zavieruka/video-platform/backend/internal/models"
)

const thumbnailSheetName = "thumbnail-candidates0000000000.jpeg"

type ThumbnailStorage interface {
	GenerateSignedDownloadURL(ctx context.Context, objectName string, expiry time.Duration) (string, error)
}

type ThumbnailHandler struct {
	videos       VideoLookup
	storage      ThumbnailStorage
	signedURLTTL time.Duration
}

func NewThumbnailHandler(videos VideoLookup, storage ThumbnailStorage, signedURLTTL time.Duration) *ThumbnailHandler {
	return &ThumbnailHandler{videos: videos, storage: storage, signedURLTTL: signedURLTTL}
}

func (h *ThumbnailHandler) ServeThumbnail(w http.ResponseWriter, r *http.Request) {
	h.serveThumbnail(w, r, false)
}

func (h *ThumbnailHandler) ServeThumbnailCandidates(w http.ResponseWriter, r *http.Request) {
	h.serveThumbnail(w, r, true)
}

func (h *ThumbnailHandler) serveThumbnail(w http.ResponseWriter, r *http.Request, candidates bool) {
	videoID := r.PathValue("id")
	if videoID == "" {
		h.respondError(w, errors.NewBadRequestError("Video ID is required"))
		return
	}

	video, err := h.videos.GetVideo(r.Context(), videoID)
	if err != nil {
		h.respondError(w, err)
		return
	}
	if video.Status != models.StatusReady {
		h.respondError(w, errors.NewConflictError("Video is not ready for thumbnail display"))
		return
	}
	if video.ThumbnailURL == nil {
		h.respondError(w, errors.NewNotFoundError("Thumbnail", videoID))
		return
	}

	objectName := videoID + "/" + thumbnailSheetName
	if !candidates && video.ThumbnailObjectName != nil {
		objectName = *video.ThumbnailObjectName
	}

	signedURL, err := h.storage.GenerateSignedDownloadURL(r.Context(), objectName, h.signedURLTTL)
	if err != nil {
		h.respondError(w, err)
		return
	}

	if candidates {
		w.Header().Set("Cache-Control", "private, max-age=60")
	} else {
		w.Header().Set("Cache-Control", "no-store")
	}
	http.Redirect(w, r, signedURL, http.StatusFound)
}

func (h *ThumbnailHandler) respondError(w http.ResponseWriter, err error) {
	appErr, ok := err.(*errors.AppError)
	if !ok {
		appErr = errors.NewInternalError("An unexpected error occurred", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.StatusCode)
	if encodeErr := json.NewEncoder(w).Encode(appErr); encodeErr != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
