package mocks

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/zavieruka/video-platform/backend/internal/models"
)

type MockVideoRepository struct {
	mock.Mock
}

func (m *MockVideoRepository) Create(ctx context.Context, video *models.Video) error {
	args := m.Called(ctx, video)
	return args.Error(0)
}

func (m *MockVideoRepository) GetByID(ctx context.Context, id string) (*models.Video, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Video), args.Error(1)
}

func (m *MockVideoRepository) UpdateStatus(ctx context.Context, id string, status models.VideoStatus, errorMsg *string) error {
	args := m.Called(ctx, id, status, errorMsg)
	return args.Error(0)
}

func (m *MockVideoRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockVideoRepository) List(ctx context.Context, limit int, offset int) ([]*models.Video, int, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.Video), args.Int(1), args.Error(2)
}

func (m *MockVideoRepository) UpdateProcessingJobID(ctx context.Context, id string, jobID string) error {
	args := m.Called(ctx, id, jobID)
	return args.Error(0)
}

func (m *MockVideoRepository) UpdateProcessingStatus(ctx context.Context, id string, status models.VideoStatus, startedAt, endedAt *time.Time) error {
	args := m.Called(ctx, id, status, startedAt, endedAt)
	return args.Error(0)
}

func (m *MockVideoRepository) UpdateProcessedVideos(ctx context.Context, id string, processedVideos map[string]models.ProcessedVideo, manifestURL, thumbnailURL string, thumbnailSelectedIndex int, endedAt *time.Time) error {
	args := m.Called(ctx, id, processedVideos, manifestURL, thumbnailURL, thumbnailSelectedIndex, endedAt)
	return args.Error(0)
}

func (m *MockVideoRepository) UpdateThumbnailSelection(ctx context.Context, id string, selectedIndex int) error {
	args := m.Called(ctx, id, selectedIndex)
	return args.Error(0)
}

func (m *MockVideoRepository) UpdateCustomThumbnail(ctx context.Context, id, objectName string) error {
	args := m.Called(ctx, id, objectName)
	return args.Error(0)
}
