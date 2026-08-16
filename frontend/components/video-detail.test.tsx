import { afterEach, describe, expect, it, vi } from "vitest";
import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

const getVideo = vi.hoisted(() => vi.fn());
const selectThumbnail = vi.hoisted(() => vi.fn());
const requestThumbnailUploadUrl = vi.hoisted(() => vi.fn());
const uploadFile = vi.hoisted(() => vi.fn());
const confirmThumbnailUpload = vi.hoisted(() => vi.fn());

vi.mock("../lib/api", () => ({
  getVideo,
  selectThumbnail,
  requestThumbnailUploadUrl,
  uploadFile,
  confirmThumbnailUpload,
  resolveApiUrl: (url: string) => url,
}));
vi.mock("./hls-player", () => ({ HlsPlayer: () => <div>player</div> }));

import { VideoDetail } from "./video-detail";

const pendingVideo = {
  id: "video-123",
  title: "Demo",
  description: "",
  fileName: "demo.mp4",
  fileSize: 1024,
  mimeType: "video/mp4",
  status: "processing" as const,
  publicUrl: "",
  createdAt: "2026-08-05T15:00:00Z",
  updatedAt: "2026-08-05T15:00:00Z",
  durationSeconds: 0,
};

describe("VideoDetail", () => {
  afterEach(() => {
    vi.useRealTimers();
    getVideo.mockReset();
    selectThumbnail.mockReset();
    requestThumbnailUploadUrl.mockReset();
    uploadFile.mockReset();
    confirmThumbnailUpload.mockReset();
  });

  it("polls nonterminal videos, then stops once the video is ready", async () => {
    vi.useFakeTimers();
    getVideo.mockResolvedValueOnce(pendingVideo).mockResolvedValueOnce({
      ...pendingVideo,
      status: "ready",
      manifestUrl: "/api/v1/videos/video-123/hls/manifest.m3u8",
    });

    render(<VideoDetail videoId="video-123" />);

    await act(async () => {
      await Promise.resolve();
    });
    expect(getVideo).toHaveBeenCalledTimes(1);
    expect(screen.getByText("Processing")).toBeInTheDocument();

    await act(async () => {
      await vi.advanceTimersByTimeAsync(3000);
    });

    expect(getVideo).toHaveBeenCalledTimes(2);
    expect(screen.getByText("Ready")).toBeInTheDocument();

    await act(async () => {
      await vi.advanceTimersByTimeAsync(6000);
    });
    expect(getVideo).toHaveBeenCalledTimes(2);
  });

  it("cleans up a scheduled poll on unmount", async () => {
    vi.useFakeTimers();
    getVideo.mockResolvedValue(pendingVideo);

    const { unmount } = render(<VideoDetail videoId="video-123" />);
    await act(async () => {
      await Promise.resolve();
    });
    expect(getVideo).toHaveBeenCalledTimes(1);

    unmount();
    await act(async () => {
      await vi.advanceTimersByTimeAsync(3000);
    });

    expect(getVideo).toHaveBeenCalledTimes(1);
  });

  it("selects a generated thumbnail candidate", async () => {
    const readyVideo = {
      ...pendingVideo,
      status: "ready" as const,
      manifestUrl: "/api/v1/videos/video-123/hls/manifest.m3u8",
      thumbnail: {
        url: "/api/v1/videos/video-123/thumbnail",
        candidatesUrl: "/api/v1/videos/video-123/thumbnail",
        selectedIndex: 5,
      },
    };
    getVideo.mockResolvedValue(readyVideo);
    selectThumbnail.mockResolvedValue({
      ...readyVideo,
      thumbnail: { ...readyVideo.thumbnail, selectedIndex: 6 },
    });

    render(<VideoDetail videoId="video-123" />);

    await act(async () => {
      await Promise.resolve();
    });
    expect(screen.getByRole("img", { name: "Demo thumbnail" }).parentElement).toHaveClass("max-w-2xl");
    screen.getByRole("button", { name: "Select thumbnail candidate 7" }).click();

    await act(async () => {
      await Promise.resolve();
    });

    expect(selectThumbnail).toHaveBeenCalledWith("video-123", 6);
    expect(screen.getByRole("button", { name: "Select thumbnail candidate 7" })).toHaveAttribute("aria-pressed", "true");
  });

  it("uses fresh thumbnail URLs after uploading a custom image or selecting a generated candidate", async () => {
    const readyVideo = {
      ...pendingVideo,
      status: "ready" as const,
      manifestUrl: "/api/v1/videos/video-123/hls/manifest.m3u8",
      thumbnail: {
        url: "/api/v1/videos/video-123/thumbnail",
        candidatesUrl: "/api/v1/videos/video-123/thumbnail",
        selectedIndex: 5,
      },
    };
    const file = new File(["image bytes"], "cover.png", { type: "image/png" });
    getVideo.mockResolvedValue(readyVideo);
    requestThumbnailUploadUrl.mockResolvedValue({ uploadUrl: "https://storage.example/signed-upload" });
    uploadFile.mockResolvedValue(undefined);
    confirmThumbnailUpload.mockResolvedValue({
      ...readyVideo,
      thumbnail: {
        url: "/api/v1/videos/video-123/thumbnail",
        candidatesUrl: "/api/v1/videos/video-123/thumbnail/candidates",
      },
    });
    selectThumbnail.mockResolvedValue({
      ...readyVideo,
      thumbnail: {
        url: "/api/v1/videos/video-123/thumbnail",
        candidatesUrl: "/api/v1/videos/video-123/thumbnail",
        selectedIndex: 6,
      },
    });

    render(<VideoDetail videoId="video-123" />);
    const user = userEvent.setup();

    await act(async () => {
      await Promise.resolve();
    });
    fireEvent.change(screen.getByLabelText("Custom thumbnail image"), { target: { files: [file] } });
    await user.click(screen.getByRole("button", { name: "Upload custom thumbnail" }));

    await waitFor(() => expect(confirmThumbnailUpload).toHaveBeenCalledWith("video-123"));
    expect(requestThumbnailUploadUrl).toHaveBeenCalledWith("video-123", { mimeType: "image/png", fileSize: file.size });
    expect(uploadFile).toHaveBeenCalledWith("https://storage.example/signed-upload", file);
    expect(screen.getByRole("img", { name: "Demo thumbnail" })).toHaveAttribute(
      "src",
      "/api/v1/videos/video-123/thumbnail?thumbnailRevision=1",
    );
    expect(screen.getByRole("button", { name: "Select thumbnail candidate 1" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Select thumbnail candidate 7" }));

    await waitFor(() => expect(selectThumbnail).toHaveBeenCalledWith("video-123", 6));
    expect(screen.getByRole("img", { name: "Demo thumbnail" })).toHaveStyle({
      backgroundImage: 'url("/api/v1/videos/video-123/thumbnail?thumbnailRevision=2")',
    });
  });
});
