package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zavieruka/video-platform/backend/internal/errors"
	"github.com/zavieruka/video-platform/backend/internal/handlers"
	"github.com/zavieruka/video-platform/backend/internal/models"
)

type fakeVideoLookup struct {
	video *models.Video
	err   error
}

func (f *fakeVideoLookup) GetVideo(context.Context, string) (*models.Video, error) {
	return f.video, f.err
}

type fakeHLSStorage struct {
	files   map[string][]byte
	readErr error
}

func (f *fakeHLSStorage) ReadFile(_ context.Context, objectName string) ([]byte, error) {
	if f.readErr != nil {
		return nil, f.readErr
	}
	data, ok := f.files[objectName]
	if !ok {
		return nil, errors.NewNotFoundError("File", objectName)
	}
	return data, nil
}

func (f *fakeHLSStorage) GenerateSignedDownloadURL(_ context.Context, objectName string, _ time.Duration) (string, error) {
	return "https://signed.example/" + objectName + "?sig=x", nil
}

func readyVideo() *models.Video {
	return &models.Video{ID: "v1", Status: models.StatusReady}
}

func serve(h *handlers.HLSHandler, id, file string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.SetPathValue("id", id)
	req.SetPathValue("file", file)
	rr := httptest.NewRecorder()
	h.ServePlaylist(rr, req)
	return rr
}

func TestServePlaylist_MasterPassthrough(t *testing.T) {
	master := "#EXTM3U\n#EXT-X-STREAM-INF:BANDWIDTH=5000000\nvideo-1080p.m3u8\n"
	storage := &fakeHLSStorage{files: map[string][]byte{"v1/manifest.m3u8": []byte(master)}}
	h := handlers.NewHLSHandler(&fakeVideoLookup{video: readyVideo()}, storage, "manifest.m3u8", time.Hour)

	rr := serve(h, "v1", "manifest.m3u8")

	require.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/vnd.apple.mpegurl", rr.Header().Get("Content-Type"))
	// Master is served verbatim — no signing, relative refs intact.
	assert.Equal(t, master, rr.Body.String())
}

func TestServePlaylist_RenditionSignsSegments(t *testing.T) {
	media := strings.Join([]string{
		"#EXTM3U",
		`#EXT-X-MAP:URI="video-1080pinit.mp4"`,
		"#EXTINF:6.000,",
		"video-1080p0.m4s",
		"#EXT-X-ENDLIST",
		"",
	}, "\n")
	storage := &fakeHLSStorage{files: map[string][]byte{"v1/video-1080p.m3u8": []byte(media)}}
	h := handlers.NewHLSHandler(&fakeVideoLookup{video: readyVideo()}, storage, "manifest.m3u8", time.Hour)

	rr := serve(h, "v1", "video-1080p.m3u8")

	require.Equal(t, http.StatusOK, rr.Code)
	body := rr.Body.String()
	assert.Contains(t, body, `#EXT-X-MAP:URI="https://signed.example/v1/video-1080pinit.mp4?sig=x"`)
	assert.Contains(t, body, "https://signed.example/v1/video-1080p0.m4s?sig=x")
	assert.NotContains(t, body, "\nvideo-1080p0.m4s\n")
}

func TestServePlaylist_NotReadyReturnsConflict(t *testing.T) {
	storage := &fakeHLSStorage{files: map[string][]byte{}}
	pending := &models.Video{ID: "v1", Status: models.StatusProcessing}
	h := handlers.NewHLSHandler(&fakeVideoLookup{video: pending}, storage, "manifest.m3u8", time.Hour)

	rr := serve(h, "v1", "manifest.m3u8")

	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestServePlaylist_VideoNotFoundPropagates(t *testing.T) {
	storage := &fakeHLSStorage{files: map[string][]byte{}}
	h := handlers.NewHLSHandler(&fakeVideoLookup{err: errors.NewNotFoundError("Video", "v1")}, storage, "manifest.m3u8", time.Hour)

	rr := serve(h, "v1", "manifest.m3u8")

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestServePlaylist_NonPlaylistRejected(t *testing.T) {
	storage := &fakeHLSStorage{files: map[string][]byte{}}
	h := handlers.NewHLSHandler(&fakeVideoLookup{video: readyVideo()}, storage, "manifest.m3u8", time.Hour)

	rr := serve(h, "v1", "video-1080p0.m4s")

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestServePlaylist_PathTraversalRejected(t *testing.T) {
	storage := &fakeHLSStorage{files: map[string][]byte{}}
	h := handlers.NewHLSHandler(&fakeVideoLookup{video: readyVideo()}, storage, "manifest.m3u8", time.Hour)

	rr := serve(h, "v1", "../other.m3u8")

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestServePlaylist_MissingObjectPropagatesNotFound(t *testing.T) {
	storage := &fakeHLSStorage{files: map[string][]byte{}} // v1/manifest.m3u8 absent
	h := handlers.NewHLSHandler(&fakeVideoLookup{video: readyVideo()}, storage, "manifest.m3u8", time.Hour)

	rr := serve(h, "v1", "manifest.m3u8")

	assert.Equal(t, http.StatusNotFound, rr.Code)
}
