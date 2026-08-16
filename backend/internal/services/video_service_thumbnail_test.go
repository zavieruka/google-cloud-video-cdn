package services_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zavieruka/video-platform/backend/internal/models"
)

func TestVideoService_SelectThumbnail_ReadyVideo(t *testing.T) {
	service, mockRepo, _, _, _ := newTestVideoService()
	ctx := context.Background()
	videoID := "video-123"
	thumbnailURL := "/api/v1/videos/video-123/thumbnail"

	mockRepo.On("GetByID", ctx, videoID).Return(&models.Video{
		ID:           videoID,
		Status:       models.StatusReady,
		ThumbnailURL: &thumbnailURL,
	}, nil)
	mockRepo.On("UpdateThumbnailSelection", ctx, videoID, 11).Return(nil)

	video, err := service.SelectThumbnail(ctx, videoID, 11)

	require.NoError(t, err)
	require.NotNil(t, video.ThumbnailSelectedIndex)
	require.Equal(t, 11, *video.ThumbnailSelectedIndex)
	mockRepo.AssertExpectations(t)
}

func TestVideoService_SelectThumbnail_RejectsOutOfRangeIndex(t *testing.T) {
	service, mockRepo, _, _, _ := newTestVideoService()
	ctx := context.Background()
	videoID := "video-123"
	thumbnailURL := "/api/v1/videos/video-123/thumbnail"

	mockRepo.On("GetByID", ctx, videoID).Return(&models.Video{
		ID:           videoID,
		Status:       models.StatusReady,
		ThumbnailURL: &thumbnailURL,
	}, nil)

	_, err := service.SelectThumbnail(ctx, videoID, 12)

	require.Error(t, err)
	require.Contains(t, err.Error(), "selectedIndex must be between 0 and 11")
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "UpdateThumbnailSelection")
}
