import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";

const hls = vi.hoisted(() => ({
  isSupported: vi.fn(),
  loadSource: vi.fn(),
  attachMedia: vi.fn(),
  destroy: vi.fn(),
  on: vi.fn(),
}));

vi.mock("hls.js", () => ({
  default: class MockHls {
    static isSupported = hls.isSupported;
    static Events = { ERROR: "hlsError" };

    loadSource = hls.loadSource;
    attachMedia = hls.attachMedia;
    destroy = hls.destroy;
    on = hls.on;
  },
}));

import { HlsPlayer } from "./hls-player";

describe("HlsPlayer", () => {
  beforeEach(() => {
    vi.stubEnv("NEXT_PUBLIC_API_BASE_URL", "http://localhost:8080");
    Object.defineProperty(HTMLMediaElement.prototype, "canPlayType", {
      configurable: true,
      value: vi.fn(() => ""),
    });
  });

  afterEach(() => {
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
    hls.isSupported.mockReset();
    hls.loadSource.mockReset();
    hls.attachMedia.mockReset();
    hls.destroy.mockReset();
    hls.on.mockReset();
  });

  it("uses native HLS when the browser supports it", () => {
    Object.defineProperty(HTMLMediaElement.prototype, "canPlayType", {
      configurable: true,
      value: vi.fn(() => "probably"),
    });

    render(<HlsPlayer manifestUrl="/api/v1/videos/video-123/hls/manifest.m3u8" />);

    const video = screen.getByLabelText("Video player") as HTMLVideoElement;
    expect(video.src).toBe("http://localhost:8080/api/v1/videos/video-123/hls/manifest.m3u8");
    expect(hls.isSupported).not.toHaveBeenCalled();
  });

  it("uses hls.js and destroys it when native HLS is unavailable", () => {
    hls.isSupported.mockReturnValue(true);

    const { unmount } = render(
      <HlsPlayer manifestUrl="/api/v1/videos/video-123/hls/manifest.m3u8" />,
    );

    const video = screen.getByLabelText("Video player");
    expect(hls.loadSource).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/videos/video-123/hls/manifest.m3u8",
    );
    expect(hls.attachMedia).toHaveBeenCalledWith(video);

    unmount();

    expect(hls.destroy).toHaveBeenCalledOnce();
  });
});
