import { afterEach, describe, expect, it, vi } from "vitest";
import { act, render, screen } from "@testing-library/react";

const getVideo = vi.hoisted(() => vi.fn());

vi.mock("../lib/api", () => ({ getVideo }));
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
});
