# Video Platform Backend - Google Cloud Platform

A scalable, serverless video content distribution platform backend built with Go and Google Cloud Platform services.

## Architecture Overview

This backend follows Google Cloud's recommended architecture for serverless applications.

The system consists of a Cloud Run service handling API requests, with data persistence managed by Firestore for metadata and Cloud Storage for video files. This architecture enables automatic scaling, high availability, and cost-effective operations through serverless infrastructure.

This backend is designed as a standalone service that exposes a REST API. It can be consumed by the accompanying Next.js frontend, mobile applications, or integrated with other backend services.

### GCP Services Used

- **Cloud Run**: Serverless compute platform for containerized applications with automatic scaling
- **Firestore**: NoSQL document database for video metadata storage
- **Cloud Storage**: Object storage for video files with global availability
- **Cloud Build**: Container image building and deployment automation
- **Cloud CDN**: Content delivery network for global video distribution
- **Pub/Sub**: Asynchronous messaging for video processing workflows
- **Transcoder API**: Video transcoding service for format conversion

### Design Rationale

**Cloud Run**
Selected for its serverless architecture, which eliminates infrastructure management overhead. Cloud Run provides automatic scaling from zero to handle variable traffic patterns, built-in HTTPS and health check mechanisms, and seamless regional deployment capabilities.

**Firestore**
Chosen as the metadata store due to its serverless nature, which aligns with the overall architecture philosophy. Firestore offers automatic scaling, real-time synchronization capabilities for future features, and strong consistency guarantees suitable for document-oriented video metadata.

**Cloud Storage**
Purpose-built for large object storage such as video files. Provides seamless integration with Cloud CDN for global content delivery, lifecycle management policies for cost optimization, and eleven nines of durability.

## Project Structure

```
backend/
├── cmd/
│   └── api/              # Main API server entrypoint
│       └── main.go
├── internal/             # Private application code
│   ├── config/          # Configuration management
│   ├── handlers/        # HTTP handlers (controllers)
│   ├── services/        # Business logic layer
│   ├── storage/         # GCS operations
│   ├── database/        # Firestore operations
│   └── middleware/      # HTTP middleware (auth, logging, etc.)
├── pkg/                 # Public libraries (if any)
├── .env.example         # Environment variables template
├── .gitignore
├── go.mod
├── go.sum
├── Dockerfile
└── README.md
```

## Prerequisites

### Local Development

- Go 1.26 or later
- Google Cloud SDK (`gcloud` CLI)
- A GCP project with billing enabled
- Docker (for local container testing)

### GCP Setup

1. **Install and authenticate gcloud CLI:**

   ```bash
   # Install: https://cloud.google.com/sdk/docs/install

   # Authenticate
   gcloud auth login
   gcloud auth application-default login

   # List projects to check ID
   gcloud projects list

   # Set your project
   gcloud config set project YOUR_PROJECT_ID
   ```

2. **Enable required APIs:**

   ```bash
   gcloud services enable \
     run.googleapis.com \
     firestore.googleapis.com \
     storage.googleapis.com \
     cloudbuild.googleapis.com
   ```

3. **Create Firestore database:**

   **IMPORTANT**: If you already have a Firestore database in your project or want to use a non-default database name, see the configuration section below.

   ```bash
   # Check if you already have Firestore databases
   gcloud firestore databases list

   # If no databases exist, create the default database
   # Create in Native mode (required for this app)
   gcloud firestore databases create --location=us-central1

   # Optional: Create a named database for this project
   # gcloud firestore databases create --database=video-platform --location=us-central1
   ```

   **Multiple Database Scenario**: If your GCP project uses Firestore for multiple applications, you can create a dedicated database for this project and specify its name in the `FIRESTORE_DATABASE_ID` environment variable. The application defaults to `(default)` if not specified.

4. **Create Cloud Storage buckets:**

   ```bash
   export PROJECT_ID=$(gcloud config get-value project)

   # Source videos bucket (user uploads)
   gsutil mb -l us-central1 gs://${PROJECT_ID}-videos-source

   # Processed videos bucket (transcoded - future)
   gsutil mb -l us-central1 gs://${PROJECT_ID}-videos-processed

   # Enable uniform bucket-level access (recommended)
   gsutil uniformbucketlevelaccess set on gs://${PROJECT_ID}-videos-source
   gsutil uniformbucketlevelaccess set on gs://${PROJECT_ID}-videos-processed
   ```

## Configuration

### Environment Variables

Copy `.env.example` to `.env` and fill in your values:

```bash
cp .env.example .env
```

Required variables:

- `GCP_PROJECT_ID`: Your GCP project ID
- `GCP_REGION`: Deployment region (e.g., us-central1)
- `SOURCE_BUCKET_NAME`: Cloud Storage bucket for uploads
- `PROCESSED_BUCKET_NAME`: Cloud Storage bucket for processed videos
- `PORT`: API server port (default: 8080)
- `ENVIRONMENT`: dev/staging/production

Optional variables:

- `FIRESTORE_DATABASE_ID`: Firestore database name (default: "(default)")
- `LOG_LEVEL`: Logging level (default: "info")

### Production Deployment

In Cloud Run, environment variables should be set via:

```bash
gcloud run deploy --set-env-vars KEY=VALUE
```

or through the Cloud Console environment variables panel.

**Do not upload a .env file to production.**

### Configuration Validation

The application validates all required configuration on startup and fails fast with clear error messages if misconfigured.

## Getting Started

### 1. Clone and Install Dependencies

```bash
cd backend
go mod download
```

### 2. Set Up Environment

```bash
cp .env.example .env
# Edit .env with your values
```

### 3. Run Locally

```bash
go run cmd/api/main.go
```

The API will start on `http://localhost:8080` (or your configured PORT).

### 4. Test the Health Check

```bash
curl http://localhost:8080/health
```

Expected response:

```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "0.1.0",
  "environment": "dev"
}
```

## API Endpoints

### Health & Status

- `GET /health` - Health check endpoint
- `GET /ready` - Readiness probe (checks GCP client connections)

### Videos (Planned)

- `POST /api/v1/videos` - Upload new video
- `GET /api/v1/videos` - List videos (paginated)
- `GET /api/v1/videos/{id}` - Get video details
- `PUT /api/v1/videos/{id}` - Update video metadata
- `DELETE /api/v1/videos/{id}` - Delete video

## Architecture Decisions

### Standalone REST API Design

This backend is intentionally decoupled from any specific frontend implementation. It exposes standard REST endpoints that can be consumed by:

- The Next.js frontend in this repository
- Mobile applications (iOS, Android)
- Other backend services
- Third-party integrations
- CLI tools

All responses follow consistent JSON formatting and standard HTTP status codes.

### Internal Package Structure

Following Go best practices, we use `internal/` for private application code that shouldn't be imported by other projects. This enforces encapsulation.

### Service Layer Pattern

Separating business logic from HTTP handlers:

- Easier testing (mock services vs handlers)
- Reusable logic across different handlers
- Clearer separation of concerns
- Follows Clean Architecture principles

### Centralized Configuration

All configuration is managed through the `config` package:

- Single source of truth
- Validation at startup (fail fast)
- Easy to test different configurations
- Type-safe access to config values

### Error Handling Strategy

- Use custom error types for different scenarios
- Wrap errors with context using `fmt.Errorf` with `%w`
- Log errors with appropriate severity
- Return user-friendly error messages without exposing internals

## Best Practices Implemented

### 12-Factor App Principles

- Configuration via environment variables
- Explicit dependencies (go.mod)
- Stateless processes
- Logs to stdout

### Google Cloud Best Practices

- Service accounts with minimal permissions
- Regional deployment (can expand to multi-region)
- Structured logging for Cloud Logging
- Health check endpoints for Cloud Run
- Support for multiple Firestore databases

### Go Best Practices

- Standard project layout
- Context propagation
- Proper error handling
- Clear package organization

## Cost Optimization Considerations

- **Cloud Run**: Scales to zero when not in use
- **Firestore**: Optimize queries to minimize document reads/writes
- **Cloud Storage**: Use lifecycle policies (Nearline/Coldline for old videos)
- **CDN**: Reduces egress costs (planned implementation)

## Security Considerations

- Never commit `.env` file (included in .gitignore)
- Use service accounts with minimal IAM permissions
- Enable uniform bucket-level access on Cloud Storage buckets
- Validate all user inputs
- Use signed URLs for time-limited access (planned)
- HTTPS only (enforced by Cloud Run)
- Support for named Firestore databases prevents conflicts

## Contributing

This is a learning project following GCP Professional Cloud Architect best practices. Contributions are welcome through issues and pull requests.

## Resources

- [Google Cloud Architecture Center](https://cloud.google.com/architecture)
- [Cloud Run Best Practices](https://cloud.google.com/run/docs/tips/general)
- [Firestore Best Practices](https://cloud.google.com/firestore/docs/best-practices)
- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [Firestore Multi-Database Support](https://cloud.google.com/firestore/docs/manage-databases)