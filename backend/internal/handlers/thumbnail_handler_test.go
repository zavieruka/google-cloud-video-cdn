package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zavieruka/video-platform/backend/internal/handlers"
	"github.com/zavieruka/video-platform/backend/internal/models"
)

func TestServeThumbnail_RedirectsReadyVideoToSignedSheet(t *testing.T) {
	thumbnailURL := "/api/v1/videos/v1/thumbnail"
	storage := &fakeHLSStorage{}
	h := handlers.NewThumbnailHandler(
		&fakeVideoLookup{video: &models.Video{ID: "v1", Status: models.StatusReady, ThumbnailURL: &thumbnailURL}},
		storage,
		time.Hour,
	)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/videos/v1/thumbnail", nil)
	req.SetPathValue("id", "v1")
	rr := httptest.NewRecorder()

	h.ServeThumbnail(rr, req)

	require.Equal(t, http.StatusFound, rr.Code)
	assert.Equal(t, "https://signed.example/v1/thumbnail-candidates0000000000.jpeg?sig=x", rr.Header().Get("Location"))
	assert.Equal(t, []string{"v1/thumbnail-candidates0000000000.jpeg"}, storage.signedObjectNames)
	assert.Contains(t, rr.Header().Get("Cache-Control"), "private")
}

func TestServeThumbnail_RejectsVideoWithoutGeneratedThumbnail(t *testing.T) {
	storage := &fakeHLSStorage{}
	h := handlers.NewThumbnailHandler(
		&fakeVideoLookup{video: &models.Video{ID: "v1", Status: models.StatusReady}},
		storage,
		time.Hour,
	)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/videos/v1/thumbnail", nil)
	req.SetPathValue("id", "v1")
	rr := httptest.NewRecorder()

	h.ServeThumbnail(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Empty(t, storage.signedObjectNames)
}
