export type VideoStatus = "pending" | "uploaded" | "processing" | "ready" | "failed";

export interface UploadURLRequest {
  title: string;
  description: string;
  fileName: string;
  fileSize: number;
  mimeType: string;
}

export interface UploadURLMetadata {
  title: string;
  description: string;
  fileName: string;
  fileSize: number;
  mimeType: string;
  objectName: string;
  status: VideoStatus;
}

export interface UploadURLResponse {
  videoId: string;
  uploadUrl: string;
  expiresAt: string;
  metadata: UploadURLMetadata;
}

export interface ProcessingStatus {
  jobId: string;
  startedAt?: string;
  endedAt?: string;
  durationSeconds?: number;
}

export interface ProcessedVideo {
  resolution: string;
  url: string;
  fileSize: number;
  bitrate: number;
}

export interface Video {
  id: string;
  title: string;
  description: string;
  fileName: string;
  fileSize: number;
  mimeType: string;
  status: VideoStatus;
  publicUrl: string;
  createdAt: string;
  updatedAt: string;
  lastError?: string;
  processingStatus?: ProcessingStatus;
  processedVideos?: ProcessedVideo[];
  manifestUrl?: string;
  durationSeconds: number;
}

export interface VideoListResponse {
  videos: Video[];
  totalCount: number;
  limit: number;
  offset: number;
}

export interface ApiErrorPayload {
  error: string;
  message: string;
  details?: Record<string, unknown>;
}

export interface FailUploadRequest {
  error: string;
  message: string;
}

export interface FailUploadResponse {
  id: string;
  status: VideoStatus;
  message: string;
}
