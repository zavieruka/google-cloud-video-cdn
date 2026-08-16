import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
  ApiError,
  confirmThumbnailUpload,
  confirmUpload,
  deleteVideo,
  getVideo,
  listVideos,
  requestThumbnailUploadUrl,
  requestUploadUrl,
  resolveApiUrl,
} from "./api";

describe("API client", () => {
  beforeEach(() => {
    vi.stubEnv("NEXT_PUBLIC_API_BASE_URL", "http://localhost:8080/");
  });

  afterEach(() => {
    vi.unstubAllEnvs();
    vi.unstubAllGlobals();
  });

  it("resolves an API-relative URL against the configured API base", () => {
    expect(resolveApiUrl("/api/v1/videos/video-123/hls/manifest.m3u8")).toBe(
      "http://localhost:8080/api/v1/videos/video-123/hls/manifest.m3u8",
    );
  });

  it("requests a signed upload URL with the backend's exact wire shape", async () => {
    const response = {
      videoId: "video-123",
      uploadUrl: "https://storage.googleapis.com/signed-upload",
      expiresAt: "2026-08-05T15:00:00Z",
      metadata: {
        title: "Demo",
        description: "A short test video",
        fileName: "demo.mp4",
        fileSize: 1024,
        mimeType: "video/mp4",
        objectName: "videos/video-123.mp4",
        status: "pending",
      },
    };
    const fetchMock = vi.fn().mockResolvedValue(Response.json(response));
    vi.stubGlobal("fetch", fetchMock);

    await expect(
      requestUploadUrl({
        title: "Demo",
        description: "A short test video",
        fileName: "demo.mp4",
        fileSize: 1024,
        mimeType: "video/mp4",
      }),
    ).resolves.toEqual(response);

    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/videos/upload-url",
      {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          title: "Demo",
          description: "A short test video",
          fileName: "demo.mp4",
          fileSize: 1024,
          mimeType: "video/mp4",
        }),
        signal: undefined,
      },
    );
  });

  it("preserves the API's status, code, message, and details on failure", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        Response.json(
          {
            error: "conflict",
            message: "Video is not ready for playback",
            details: { status: "processing" },
          },
          { status: 409 },
        ),
      ),
    );

    let caught: unknown;
    try {
      await requestUploadUrl({
        title: "Demo",
        description: "",
        fileName: "demo.mp4",
        fileSize: 1024,
        mimeType: "video/mp4",
      });
    } catch (error) {
      caught = error;
    }

    expect(caught).toBeInstanceOf(ApiError);
    expect(caught).toMatchObject({
      name: "ApiError",
      status: 409,
      code: "conflict",
      message: "Video is not ready for playback",
      details: { status: "processing" },
    });
  });

  it("uses the confirmation, list, detail, and delete endpoints", async () => {
    const video = {
      id: "video-123",
      title: "Demo",
      description: "",
      fileName: "demo.mp4",
      fileSize: 1024,
      mimeType: "video/mp4",
      status: "uploaded",
      publicUrl: "",
      createdAt: "2026-08-05T15:00:00Z",
      updatedAt: "2026-08-05T15:00:00Z",
      durationSeconds: 0,
    };
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(Response.json(video))
      .mockResolvedValueOnce(Response.json({ videos: [video], totalCount: 1, limit: 5, offset: 10 }))
      .mockResolvedValueOnce(Response.json(video))
      .mockResolvedValueOnce(Response.json({ message: "Video deleted successfully" }));
    vi.stubGlobal("fetch", fetchMock);

    await confirmUpload("video-123", "2026-08-05T15:05:00Z");
    await listVideos({ limit: 5, offset: 10 });
    await getVideo("video-123");
    await deleteVideo("video-123");

    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      "http://localhost:8080/api/v1/videos/video-123/confirm",
      {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ uploadedAt: "2026-08-05T15:05:00Z" }),
        signal: undefined,
      },
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      "http://localhost:8080/api/v1/videos?limit=5&offset=10",
      { signal: undefined },
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      3,
      "http://localhost:8080/api/v1/videos/video-123",
      { signal: undefined },
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      4,
      "http://localhost:8080/api/v1/videos/video-123",
      { method: "DELETE", headers: undefined, body: undefined, signal: undefined },
    );
  });

  it("uses the custom-thumbnail signed upload and confirmation endpoints", async () => {
    const uploadResponse = {
      uploadUrl: "https://storage.googleapis.com/signed-thumbnail-upload",
      expiresAt: "2026-08-05T15:00:00Z",
    };
    const video = {
      id: "video-123",
      title: "Demo",
      description: "",
      fileName: "demo.mp4",
      fileSize: 1024,
      mimeType: "video/mp4",
      status: "ready",
      publicUrl: "",
      createdAt: "2026-08-05T15:00:00Z",
      updatedAt: "2026-08-05T15:00:00Z",
      durationSeconds: 0,
    };
    const fetchMock = vi.fn().mockResolvedValueOnce(Response.json(uploadResponse)).mockResolvedValueOnce(Response.json(video));
    vi.stubGlobal("fetch", fetchMock);

    await expect(requestThumbnailUploadUrl("video-123", { mimeType: "image/png", fileSize: 1024 })).resolves.toEqual(uploadResponse);
    await expect(confirmThumbnailUpload("video-123")).resolves.toEqual(video);

    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      "http://localhost:8080/api/v1/videos/video-123/thumbnail/upload-url",
      {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ mimeType: "image/png", fileSize: 1024 }),
        signal: undefined,
      },
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      "http://localhost:8080/api/v1/videos/video-123/thumbnail/confirm",
      { method: "POST", headers: undefined, body: undefined, signal: undefined },
    );
  });
});
