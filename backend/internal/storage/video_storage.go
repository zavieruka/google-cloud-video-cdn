package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	iamcredentials "cloud.google.com/go/iam/credentials/apiv1"
	credentialspb "google.golang.org/genproto/googleapis/iam/credentials/v1"

	"cloud.google.com/go/storage"
	"github.com/zavieruka/video-platform/backend/internal/errors"
	"google.golang.org/api/iterator"
)

type VideoStorage interface {
	GenerateSignedUploadURL(ctx context.Context, objectName string, mimeType string, expiryDuration time.Duration) (string, error)
	FileExists(ctx context.Context, objectName string) (bool, error)
	GetFileSize(ctx context.Context, objectName string) (int64, error)
	DeleteFile(ctx context.Context, objectName string) error
	DeleteByPrefix(ctx context.Context, prefix string) error
	GetPublicURL(objectName string) string
	GetStorageURL(objectName string) string
}

type GCSVideoStorage struct {
	client              *storage.Client
	bucketName          string
	serviceAccountEmail string
}

func NewGCSVideoStorage(client *storage.Client, bucketName string, serviceAccountEmail string) *GCSVideoStorage {
	return &GCSVideoStorage{
		client:              client,
		bucketName:          bucketName,
		serviceAccountEmail: serviceAccountEmail,
	}
}

func (s *GCSVideoStorage) GenerateSignedUploadURL(
	ctx context.Context,
	objectName string,
	mimeType string,
	expiryDuration time.Duration,
) (string, error) {
	return s.signedURL(ctx, objectName, http.MethodPut, mimeType, expiryDuration)
}

// GenerateSignedDownloadURL returns a time-limited V4 signed URL that grants
// read access to a single object in this bucket. It lets the HLS delivery layer
// hand out segment URLs while keeping the processed bucket private. No
// Content-Type is signed: a GET request must not send one for the signature to
// match.
func (s *GCSVideoStorage) GenerateSignedDownloadURL(
	ctx context.Context,
	objectName string,
	expiryDuration time.Duration,
) (string, error) {
	return s.signedURL(ctx, objectName, http.MethodGet, "", expiryDuration)
}

// ReadFile returns the full contents of an object. It is used to serve the
// small HLS playlists (master + rendition) through the API; segments are never
// read this way, they are fetched directly from GCS via signed URLs.
func (s *GCSVideoStorage) ReadFile(ctx context.Context, objectName string) ([]byte, error) {
	reader, err := s.client.Bucket(s.bucketName).Object(objectName).NewReader(ctx)
	if err == storage.ErrObjectNotExist {
		return nil, errors.NewNotFoundError("File", objectName)
	}
	if err != nil {
		return nil, errors.NewStorageError("Failed to open file for reading", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.NewStorageError("Failed to read file", err)
	}

	return data, nil
}

// signedURL builds a V4 signed URL for objectName, signing the blob with the
// configured service account via the IAM Credentials API (no private key on
// disk). It backs both the upload (PUT) and download (GET) URL generators.
func (s *GCSVideoStorage) signedURL(
	ctx context.Context,
	objectName string,
	method string,
	contentType string,
	expiryDuration time.Duration,
) (string, error) {
	iamClient, err := iamcredentials.NewIamCredentialsClient(ctx)
	if err != nil {
		return "", err
	}
	defer iamClient.Close()

	signBytes := func(b []byte) ([]byte, error) {
		req := &credentialspb.SignBlobRequest{
			Name:    "projects/-/serviceAccounts/" + s.serviceAccountEmail,
			Payload: b,
		}

		resp, err := iamClient.SignBlob(ctx, req)
		if err != nil {
			return nil, err
		}

		return resp.SignedBlob, nil
	}

	opts := &storage.SignedURLOptions{
		Scheme:         storage.SigningSchemeV4,
		Method:         method,
		Expires:        time.Now().Add(expiryDuration),
		ContentType:    contentType,
		GoogleAccessID: s.serviceAccountEmail,
		SignBytes:      signBytes,
	}

	url, err := storage.SignedURL(s.bucketName, objectName, opts)
	if err != nil {
		return "", err
	}

	return url, nil
}

func (s *GCSVideoStorage) FileExists(ctx context.Context, objectName string) (bool, error) {
	bucket := s.client.Bucket(s.bucketName)
	object := bucket.Object(objectName)

	_, err := object.Attrs(ctx)
	if err == storage.ErrObjectNotExist {
		return false, nil
	}
	if err != nil {
		return false, errors.NewStorageError("Failed to check file existence", err)
	}

	return true, nil
}

func (s *GCSVideoStorage) GetFileSize(ctx context.Context, objectName string) (int64, error) {
	bucket := s.client.Bucket(s.bucketName)
	object := bucket.Object(objectName)

	attrs, err := object.Attrs(ctx)
	if err == storage.ErrObjectNotExist {
		return 0, errors.NewNotFoundError("File", objectName)
	}
	if err != nil {
		return 0, errors.NewStorageError("Failed to get file size", err)
	}

	return attrs.Size, nil
}

func (s *GCSVideoStorage) DeleteFile(ctx context.Context, objectName string) error {
	bucket := s.client.Bucket(s.bucketName)
	object := bucket.Object(objectName)

	err := object.Delete(ctx)
	if err != nil && err != storage.ErrObjectNotExist {
		return errors.NewStorageError("Failed to delete file", err)
	}

	return nil
}

// DeleteByPrefix deletes every object under prefix in this bucket. It is used to
// purge a video's processed HLS output (all objects under "{videoID}/") when the
// video is deleted, so transcoded artifacts don't outlive the record. Objects
// that vanish mid-iteration are ignored; the first hard error aborts.
func (s *GCSVideoStorage) DeleteByPrefix(ctx context.Context, prefix string) error {
	bucket := s.client.Bucket(s.bucketName)
	it := bucket.Objects(ctx, &storage.Query{Prefix: prefix})

	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return errors.NewStorageError("Failed to list objects for deletion", err)
		}

		if err := bucket.Object(attrs.Name).Delete(ctx); err != nil && err != storage.ErrObjectNotExist {
			return errors.NewStorageError("Failed to delete object", err)
		}
	}

	return nil
}

// GetPublicURL returns the public URL for accessing a file
func (s *GCSVideoStorage) GetPublicURL(objectName string) string {
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", s.bucketName, objectName)
}

// GetStorageURL returns the gs:// URL for a file
func (s *GCSVideoStorage) GetStorageURL(objectName string) string {
	return fmt.Sprintf("gs://%s/%s", s.bucketName, objectName)
}
