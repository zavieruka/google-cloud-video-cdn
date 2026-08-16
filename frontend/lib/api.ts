import type {
  ApiErrorPayload,
  FailUploadRequest,
  FailUploadResponse,
  UploadURLRequest,
  UploadURLResponse,
  Video,
  VideoListResponse,
} from "./types";

export class ApiError extends Error {
  readonly status: number;
  readonly code: string;
  readonly details?: Record<string, unknown>;

  constructor(status: number, code: string, message: string, details?: Record<string, unknown>) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
    this.details = details;
  }
}

function getApiBaseUrl(): string {
  const baseUrl = process.env.NEXT_PUBLIC_API_BASE_URL;
  if (!baseUrl) {
    throw new Error("NEXT_PUBLIC_API_BASE_URL is required");
  }

  return baseUrl.replace(/\/+$/, "");
}

export function resolveApiUrl(path: string): string {
  if (/^https?:\/\//.test(path)) {
    return path;
  }

  const normalizedPath = path.startsWith("/") ? path : `/${path}`;
  return `${getApiBaseUrl()}${normalizedPath}`;
}

async function decodeError(response: Response): Promise<ApiError> {
  let payload: Partial<ApiErrorPayload> = {};
  try {
    payload = (await response.json()) as Partial<ApiErrorPayload>;
  } catch {
    // The API normally sends JSON errors, but retain a useful fallback for a
    // network intermediary that returns an empty or non-JSON response.
  }

  return new ApiError(
    response.status,
    payload.error ?? "request_failed",
    payload.message ?? `Request failed with status ${response.status}`,
    payload.details,
  );
}

async function requestJSON<T>(
  path: string,
  init: RequestInit,
): Promise<T> {
  const response = await fetch(resolveApiUrl(path), init);
  if (!response.ok) {
    throw await decodeError(response);
  }

  return (await response.json()) as T;
}

function jsonRequest(method: "POST" | "PATCH" | "DELETE", body?: unknown, signal?: AbortSignal): RequestInit {
  return {
    method,
    headers: body === undefined ? undefined : { "Content-Type": "application/json" },
    body: body === undefined ? undefined : JSON.stringify(body),
    signal,
  };
}

export function requestUploadUrl(
  request: UploadURLRequest,
  signal?: AbortSignal,
): Promise<UploadURLResponse> {
  return requestJSON<UploadURLResponse>(
    "/api/v1/videos/upload-url",
    jsonRequest("POST", request, signal),
  );
}

export function confirmUpload(
  videoId: string,
  uploadedAt?: string,
  signal?: AbortSignal,
): Promise<Video> {
  return requestJSON<Video>(
    `/api/v1/videos/${videoId}/confirm`,
    jsonRequest("POST", uploadedAt ? { uploadedAt } : {}, signal),
  );
}

export function failUpload(
  videoId: string,
  request: FailUploadRequest,
  signal?: AbortSignal,
): Promise<FailUploadResponse> {
  return requestJSON<FailUploadResponse>(
    `/api/v1/videos/${videoId}/fail`,
    jsonRequest("POST", request, signal),
  );
}

export function getVideo(videoId: string, signal?: AbortSignal): Promise<Video> {
  return requestJSON<Video>(`/api/v1/videos/${videoId}`, { signal });
}

export function selectThumbnail(videoId: string, selectedIndex: number, signal?: AbortSignal): Promise<Video> {
  return requestJSON<Video>(
    `/api/v1/videos/${videoId}/thumbnail`,
    jsonRequest("PATCH", { selectedIndex }, signal),
  );
}

export function listVideos(
  { limit = 20, offset = 0 }: { limit?: number; offset?: number } = {},
  signal?: AbortSignal,
): Promise<VideoListResponse> {
  const query = new URLSearchParams({ limit: String(limit), offset: String(offset) });
  return requestJSON<VideoListResponse>(`/api/v1/videos?${query}`, { signal });
}

export function deleteVideo(videoId: string, signal?: AbortSignal): Promise<{ message: string }> {
  return requestJSON<{ message: string }>(
    `/api/v1/videos/${videoId}`,
    jsonRequest("DELETE", undefined, signal),
  );
}

export type UploadProgressCallback = (progress: number) => void;

// GCS signed uploads need a browser XHR rather than fetch so the UI can expose
// meaningful byte progress. The signed URL binds this exact Content-Type.
export function uploadFile(
  uploadUrl: string,
  file: File,
  onProgress?: UploadProgressCallback,
): Promise<void> {
  return new Promise((resolve, reject) => {
    const request = new XMLHttpRequest();
    request.open("PUT", uploadUrl, true);
    request.setRequestHeader("Content-Type", file.type);

    request.upload.onprogress = (event) => {
      if (event.lengthComputable && event.total > 0) {
        onProgress?.(event.loaded / event.total);
      }
    };

    request.onload = () => {
      if (request.status >= 200 && request.status < 300) {
        resolve();
        return;
      }
      reject(new Error(`Upload failed with status ${request.status}`));
    };
    request.onerror = () => reject(new Error("Upload failed due to a network error"));
    request.onabort = () => reject(new Error("Upload was cancelled"));
    request.send(file);
  });
}
