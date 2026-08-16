package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/zavieruka/video-platform/backend/internal/models"
)

func TestVideoService_RequestThumbnailUploadURL_ReadyVideo(t *testing.T) {
	service, mockRepo, _, mockProcessed := newTestVideoServiceWithProcessed()
	ctx := context.Background()
	videoID := "video-123"
	thumbnailURL := "/api/v1/videos/video-123/thumbnail"

	mockRepo.On("GetByID", ctx, videoID).Return(&models.Video{
		ID:           videoID,
		Status:       models.StatusReady,
		ThumbnailURL: &thumbnailURL,
	}, nil)
	mockProcessed.On("GenerateSignedUploadURL", ctx, "video-123/thumbnail-upload", "image/jpeg", time.Hour).
		Return("https://storage.example/signed-upload", nil)

	response, err := service.RequestThumbnailUploadURL(ctx, videoID, &models.ThumbnailUploadURLRequest{
		MimeType: "image/jpeg",
		FileSize: 1024,
	})

	require.NoError(t, err)
	require.Equal(t, "https://storage.example/signed-upload", response.UploadURL)
	require.False(t, response.ExpiresAt.IsZero())
	mockRepo.AssertExpectations(t)
	mockProcessed.AssertExpectations(t)
}

func TestVideoService_RequestThumbnailUploadURL_RejectsUnsupportedOrOversizedImage(t *testing.T) {
	service, mockRepo, _, mockProcessed := newTestVideoServiceWithProcessed()
	ctx := context.Background()
	videoID := "video-123"
	thumbnailURL := "/api/v1/videos/video-123/thumbnail"

	mockRepo.On("GetByID", ctx, videoID).Return(&models.Video{
		ID:           videoID,
		Status:       models.StatusReady,
		ThumbnailURL: &thumbnailURL,
	}, nil)

	_, err := service.RequestThumbnailUploadURL(ctx, videoID, &models.ThumbnailUploadURLRequest{
		MimeType: "image/gif",
		FileSize: 1024,
	})
	require.ErrorContains(t, err, "JPEG, PNG, or WebP")

	_, err = service.RequestThumbnailUploadURL(ctx, videoID, &models.ThumbnailUploadURLRequest{
		MimeType: "image/png",
		FileSize: models.MaxThumbnailUploadSizeBytes + 1,
	})
	require.ErrorContains(t, err, "10 MiB")
	mockProcessed.AssertNotCalled(t, "GenerateSignedUploadURL")
}

func TestVideoService_ConfirmThumbnailUpload_SetsCustomThumbnail(t *testing.T) {
	service, mockRepo, _, mockProcessed := newTestVideoServiceWithProcessed()
	ctx := context.Background()
	videoID := "video-123"
	thumbnailURL := "/api/v1/videos/video-123/thumbnail"
	generatedIndex := 5

	mockRepo.On("GetByID", ctx, videoID).Return(&models.Video{
		ID:                     videoID,
		Status:                 models.StatusReady,
		ThumbnailURL:           &thumbnailURL,
		ThumbnailSelectedIndex: &generatedIndex,
	}, nil)
	mockProcessed.On("FileExists", ctx, "video-123/thumbnail-upload").Return(true, nil)
	mockProcessed.On("GetFileSize", ctx, "video-123/thumbnail-upload").Return(int64(1024), nil)
	mockRepo.On("UpdateCustomThumbnail", ctx, videoID, "video-123/thumbnail-upload").Return(nil)

	video, err := service.ConfirmThumbnailUpload(ctx, videoID)

	require.NoError(t, err)
	require.NotNil(t, video.ThumbnailObjectName)
	require.Equal(t, "video-123/thumbnail-upload", *video.ThumbnailObjectName)
	require.Nil(t, video.ThumbnailSelectedIndex)
	response := video.ToResponse()
	require.NotNil(t, response.Thumbnail)
	require.Equal(t, "/api/v1/videos/video-123/thumbnail", response.Thumbnail.URL)
	require.Equal(t, "/api/v1/videos/video-123/thumbnail/candidates", response.Thumbnail.CandidatesURL)
	require.Nil(t, response.Thumbnail.SelectedIndex)
	mockRepo.AssertExpectations(t)
	mockProcessed.AssertExpectations(t)
}

func TestVideoService_ConfirmThumbnailUpload_RejectsEmptyOrOversizedFile(t *testing.T) {
	for _, fileSize := range []int64{0, models.MaxThumbnailUploadSizeBytes + 1} {
		t.Run("invalid thumbnail size", func(t *testing.T) {
			service, mockRepo, _, mockProcessed := newTestVideoServiceWithProcessed()
			ctx := context.Background()
			videoID := "video-123"
			thumbnailURL := "/api/v1/videos/video-123/thumbnail"

			mockRepo.On("GetByID", ctx, videoID).Return(&models.Video{
				ID:           videoID,
				Status:       models.StatusReady,
				ThumbnailURL: &thumbnailURL,
			}, nil)
			mockProcessed.On("FileExists", ctx, "video-123/thumbnail-upload").Return(true, nil)
			mockProcessed.On("GetFileSize", ctx, "video-123/thumbnail-upload").Return(fileSize, nil)

			_, err := service.ConfirmThumbnailUpload(ctx, videoID)

			require.ErrorContains(t, err, "10 MiB")
			mockRepo.AssertNotCalled(t, "UpdateCustomThumbnail")
			mockRepo.AssertExpectations(t)
			mockProcessed.AssertExpectations(t)
		})
	}
}

func TestVideoService_SelectThumbnail_ReplacesCustomThumbnail(t *testing.T) {
	service, mockRepo, _, _ := newTestVideoServiceWithProcessed()
	ctx := context.Background()
	videoID := "video-123"
	thumbnailURL := "/api/v1/videos/video-123/thumbnail"
	customObjectName := "video-123/thumbnail-upload"

	mockRepo.On("GetByID", ctx, videoID).Return(&models.Video{
		ID:                  videoID,
		Status:              models.StatusReady,
		ThumbnailURL:        &thumbnailURL,
		ThumbnailObjectName: &customObjectName,
	}, nil)
	mockRepo.On("UpdateThumbnailSelection", ctx, videoID, 3).Return(nil)

	video, err := service.SelectThumbnail(ctx, videoID, 3)

	require.NoError(t, err)
	require.Nil(t, video.ThumbnailObjectName)
	require.NotNil(t, video.ThumbnailSelectedIndex)
	require.Equal(t, 3, *video.ThumbnailSelectedIndex)
	mockRepo.AssertExpectations(t)
}
