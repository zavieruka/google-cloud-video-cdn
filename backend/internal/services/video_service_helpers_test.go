package services_test

import (
	"fmt"
	"time"

	"github.com/zavieruka/video-platform/backend/internal/mocks"
	"github.com/zavieruka/video-platform/backend/internal/models"
	"github.com/zavieruka/video-platform/backend/internal/services"
)

func newTestVideoService() (
	*services.VideoServiceImpl,
	*mocks.MockVideoRepository,
	*mocks.MockVideoStorage,
	*mocks.MockValidator,
	*mocks.MockPublisher,
) {
	mockRepo := new(mocks.MockVideoRepository)
	mockStorage := new(mocks.MockVideoStorage)
	mockValidator := new(mocks.MockValidator)
	mockPublisher := new(mocks.MockPublisher)

	// Processed-bucket storage is only exercised by DeleteVideo on ready/failed
	// videos; the shared helper wires a fresh mock so those paths don't panic,
	// but leaves it inaccessible here. Tests that assert on it use
	// newTestVideoServiceWithProcessed.
	mockProcessed := new(mocks.MockVideoStorage)

	service := services.NewVideoService(
		mockRepo,
		mockStorage,
		mockProcessed,
		mockValidator,
		1,
		mockPublisher,
		"test-bucket-source",
		true,
	)

	return service, mockRepo, mockStorage, mockValidator, mockPublisher
}

// newTestVideoServiceWithProcessed exposes the processed-bucket storage mock, for
// tests that delete a ready/failed video and need to assert the HLS output is
// purged by prefix.
func newTestVideoServiceWithProcessed() (
	*services.VideoServiceImpl,
	*mocks.MockVideoRepository,
	*mocks.MockVideoStorage,
	*mocks.MockVideoStorage,
) {
	mockRepo := new(mocks.MockVideoRepository)
	mockStorage := new(mocks.MockVideoStorage)
	mockProcessed := new(mocks.MockVideoStorage)
	mockValidator := new(mocks.MockValidator)
	mockPublisher := new(mocks.MockPublisher)

	service := services.NewVideoService(
		mockRepo,
		mockStorage,
		mockProcessed,
		mockValidator,
		1,
		mockPublisher,
		"test-bucket-source",
		true,
	)

	return service, mockRepo, mockStorage, mockProcessed
}

// newTestVideoServiceNoPublisher builds a service with auto-processing enabled but
// no publisher configured (a genuine nil interface), mirroring cmd/api's degraded mode.
func newTestVideoServiceNoPublisher() (
	*services.VideoServiceImpl,
	*mocks.MockVideoRepository,
	*mocks.MockVideoStorage,
) {
	mockRepo := new(mocks.MockVideoRepository)
	mockStorage := new(mocks.MockVideoStorage)
	mockProcessed := new(mocks.MockVideoStorage)
	mockValidator := new(mocks.MockValidator)

	service := services.NewVideoService(
		mockRepo,
		mockStorage,
		mockProcessed,
		mockValidator,
		1,
		nil, // no publisher configured
		"test-bucket-source",
		true, // auto-processing enabled
	)

	return service, mockRepo, mockStorage
}

func newPendingVideo(videoID string) *models.Video {
	return &models.Video{
		ID:                 videoID,
		Title:              "Test Video",
		Status:             models.StatusPending,
		ObjectName:         fmt.Sprintf("videos/%s.mp4", videoID),
		FileSize:           1024 * 1024 * 10,
		MimeType:           "video/mp4",
		UploadURLExpiresAt: time.Now().UTC().Add(1 * time.Hour),
	}
}
