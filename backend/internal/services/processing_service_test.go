package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/zavieruka/video-platform/backend/internal/mocks"
	"github.com/zavieruka/video-platform/backend/internal/models"
)

// handleJobSuccess is what turns a finished transcoder job into the URLs the
// delivery layer serves. This locks in that exact URL shape: the master manifest
// and per-rendition playlists must point at the API's HLS endpoints, and the
// stored gs:// paths must match the flat layout the Transcoder template emits
// (video-1080p.m3u8, not video-1080p/media.m3u8). It also guards the flow: the
// video is only marked ready after the processed videos are recorded.
func TestProcessingService_handleJobSuccess_WritesDeliveryURLs(t *testing.T) {
	mockRepo := new(mocks.MockVideoRepository)
	svc := &ProcessingService{
		videoRepo:       mockRepo,
		processedBucket: "test-bucket-processed",
	}

	ctx := context.Background()
	videoID := "video-123"

	var gotProcessed map[string]models.ProcessedVideo
	var gotManifest string

	mockRepo.On("UpdateProcessedVideos", mock.Anything, videoID, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			gotProcessed = args.Get(2).(map[string]models.ProcessedVideo)
			gotManifest = args.String(3)
		}).Return(nil)
	mockRepo.On("UpdateStatus", mock.Anything, videoID, models.StatusReady, (*string)(nil)).Return(nil)

	svc.handleJobSuccess(ctx, videoID)

	// Both writes happened, and ready came after the processed videos were stored.
	mockRepo.AssertExpectations(t)

	assert.Equal(t, "/api/v1/videos/video-123/hls/manifest.m3u8", gotManifest)

	expected := map[string]models.ProcessedVideo{
		"1080p": {
			Resolution: "1080p",
			StorageURL: "gs://test-bucket-processed/video-123/video-1080p.m3u8",
			PublicURL:  "/api/v1/videos/video-123/hls/video-1080p.m3u8",
			Bitrate:    5000000,
		},
		"720p": {
			Resolution: "720p",
			StorageURL: "gs://test-bucket-processed/video-123/video-720p.m3u8",
			PublicURL:  "/api/v1/videos/video-123/hls/video-720p.m3u8",
			Bitrate:    2500000,
		},
		"480p": {
			Resolution: "480p",
			StorageURL: "gs://test-bucket-processed/video-123/video-480p.m3u8",
			PublicURL:  "/api/v1/videos/video-123/hls/video-480p.m3u8",
			Bitrate:    1000000,
		},
	}
	require.Equal(t, expected, gotProcessed)
}

// If recording the processed videos fails, the video must NOT be marked ready —
// otherwise a "ready" video would have no playable manifest.
func TestProcessingService_handleJobSuccess_NoReadyIfProcessedWriteFails(t *testing.T) {
	mockRepo := new(mocks.MockVideoRepository)
	svc := &ProcessingService{
		videoRepo:       mockRepo,
		processedBucket: "test-bucket-processed",
	}

	ctx := context.Background()
	videoID := "video-123"

	mockRepo.On("UpdateProcessedVideos", mock.Anything, videoID, mock.Anything, mock.Anything, mock.Anything).
		Return(assert.AnError)

	svc.handleJobSuccess(ctx, videoID)

	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "UpdateStatus")
}
