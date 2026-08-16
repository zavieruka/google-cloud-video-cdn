package database

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	firestorepb "cloud.google.com/go/firestore/apiv1/firestorepb"
	"github.com/zavieruka/video-platform/backend/internal/errors"
	"github.com/zavieruka/video-platform/backend/internal/models"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const videosCollection = "videos"

type VideoRepository interface {
	Create(ctx context.Context, video *models.Video) error
	GetByID(ctx context.Context, id string) (*models.Video, error)
	UpdateStatus(ctx context.Context, id string, status models.VideoStatus, errorMsg *string) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit int, offset int) ([]*models.Video, int, error)
	UpdateProcessingJobID(ctx context.Context, id string, jobID string) error
	UpdateProcessingStatus(ctx context.Context, id string, status models.VideoStatus, startedAt, endedAt *time.Time) error
	UpdateProcessedVideos(ctx context.Context, id string, processedVideos map[string]models.ProcessedVideo, manifestURL, thumbnailURL string, thumbnailSelectedIndex int, endedAt *time.Time) error
	UpdateThumbnailSelection(ctx context.Context, id string, selectedIndex int) error
}

type FirestoreVideoRepository struct {
	client *firestore.Client
}

func NewFirestoreVideoRepository(client *firestore.Client) *FirestoreVideoRepository {
	return &FirestoreVideoRepository{
		client: client,
	}
}

func (r *FirestoreVideoRepository) Create(ctx context.Context, video *models.Video) error {
	_, err := r.client.Collection(videosCollection).Doc(video.ID).Set(ctx, video)
	if err != nil {
		return errors.NewDatabaseError("Failed to create video", err)
	}
	return nil
}

func (r *FirestoreVideoRepository) GetByID(ctx context.Context, id string) (*models.Video, error) {
	doc, err := r.client.Collection(videosCollection).Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, errors.NewNotFoundError("Video", id)
		}
		return nil, errors.NewDatabaseError("Failed to get video", err)
	}

	var video models.Video
	if err := doc.DataTo(&video); err != nil {
		return nil, errors.NewDatabaseError("Failed to parse video data", err)
	}

	return &video, nil
}

// UpdateStatus updates the status of a video and optionally sets an error message
func (r *FirestoreVideoRepository) UpdateStatus(ctx context.Context, id string, status models.VideoStatus, errorMsg *string) error {
	updates := []firestore.Update{
		{Path: "status", Value: status},
		{Path: "updatedAt", Value: firestore.ServerTimestamp},
	}

	if errorMsg != nil {
		updates = append(updates, firestore.Update{Path: "lastError", Value: *errorMsg})
	} else {
		if status != models.StatusFailed {
			updates = append(updates, firestore.Update{Path: "lastError", Value: firestore.Delete})
		}
	}

	_, err := r.client.Collection(videosCollection).Doc(id).Update(ctx, updates)
	if err != nil {
		return errors.NewDatabaseError(fmt.Sprintf("Failed to update video status to %s", status), err)
	}

	return nil
}

func (r *FirestoreVideoRepository) Delete(ctx context.Context, id string) error {
	_, err := r.client.Collection(videosCollection).Doc(id).Delete(ctx)
	if err != nil {
		return errors.NewDatabaseError("Failed to delete video", err)
	}
	return nil
}

func (r *FirestoreVideoRepository) List(ctx context.Context, limit int, offset int) ([]*models.Video, int, error) {
	// Get total count via a server-side aggregation (avoids scanning every doc)
	aggResult, err := r.client.Collection(videosCollection).NewAggregationQuery().WithCount("all").Get(ctx)
	if err != nil {
		return nil, 0, errors.NewDatabaseError("Failed to count videos", err)
	}
	totalCount := 0
	if v, ok := aggResult["all"].(*firestorepb.Value); ok {
		totalCount = int(v.GetIntegerValue())
	}

	// Get paginated results
	query := r.client.Collection(videosCollection).
		OrderBy("createdAt", firestore.Desc).
		Limit(limit).
		Offset(offset)

	iter := query.Documents(ctx)
	defer iter.Stop()

	var videos []*models.Video
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, 0, errors.NewDatabaseError("Failed to list videos", err)
		}

		var video models.Video
		if err := doc.DataTo(&video); err != nil {
			return nil, 0, errors.NewDatabaseError("Failed to parse video data", err)
		}
		videos = append(videos, &video)
	}

	return videos, totalCount, nil
}

func (r *FirestoreVideoRepository) UpdateProcessingJobID(ctx context.Context, id string, jobID string) error {
	_, err := r.client.Collection(videosCollection).Doc(id).Update(ctx, []firestore.Update{
		{Path: "processingJobId", Value: jobID},
		{Path: "updatedAt", Value: firestore.ServerTimestamp},
	})

	if err != nil {
		return errors.NewDatabaseError("Failed to update processing job ID", err)
	}

	return nil
}

func (r *FirestoreVideoRepository) UpdateProcessingStatus(ctx context.Context, id string, status models.VideoStatus, startedAt, endedAt *time.Time) error {
	updates := []firestore.Update{
		{Path: "status", Value: status},
		{Path: "updatedAt", Value: firestore.ServerTimestamp},
	}

	if startedAt != nil {
		updates = append(updates, firestore.Update{Path: "processingStartedAt", Value: *startedAt})
	}

	if endedAt != nil {
		updates = append(updates, firestore.Update{Path: "processingEndedAt", Value: *endedAt})
	}

	_, err := r.client.Collection(videosCollection).Doc(id).Update(ctx, updates)
	if err != nil {
		return errors.NewDatabaseError("Failed to update processing status", err)
	}

	return nil
}

func (r *FirestoreVideoRepository) UpdateProcessedVideos(ctx context.Context, id string, processedVideos map[string]models.ProcessedVideo, manifestURL, thumbnailURL string, thumbnailSelectedIndex int, endedAt *time.Time) error {
	updates := []firestore.Update{
		{Path: "processedVideos", Value: processedVideos},
		{Path: "manifestUrl", Value: manifestURL},
		{Path: "thumbnailUrl", Value: thumbnailURL},
		{Path: "thumbnailSelectedIndex", Value: thumbnailSelectedIndex},
		{Path: "updatedAt", Value: firestore.ServerTimestamp},
	}

	if endedAt != nil {
		updates = append(updates, firestore.Update{Path: "processingEndedAt", Value: *endedAt})
	}

	_, err := r.client.Collection(videosCollection).Doc(id).Update(ctx, updates)

	if err != nil {
		return errors.NewDatabaseError("Failed to update processed videos", err)
	}

	return nil
}

func (r *FirestoreVideoRepository) UpdateThumbnailSelection(ctx context.Context, id string, selectedIndex int) error {
	_, err := r.client.Collection(videosCollection).Doc(id).Update(ctx, []firestore.Update{
		{Path: "thumbnailSelectedIndex", Value: selectedIndex},
		{Path: "updatedAt", Value: firestore.ServerTimestamp},
	})
	if err != nil {
		return errors.NewDatabaseError("Failed to update thumbnail selection", err)
	}

	return nil
}
