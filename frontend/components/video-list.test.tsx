import { render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

const listVideos = vi.hoisted(() => vi.fn());

vi.mock("../lib/api", () => ({ listVideos, resolveApiUrl: (url: string) => url }));

import { VideoList } from "./video-list";

describe("VideoList", () => {
  afterEach(() => {
    listVideos.mockReset();
  });

  it("renders the selected generated thumbnail for a ready video", async () => {
    listVideos.mockResolvedValue({
      videos: [
        {
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
          thumbnail: {
            url: "/api/v1/videos/video-123/thumbnail",
            candidatesUrl: "/api/v1/videos/video-123/thumbnail",
            selectedIndex: 5,
          },
        },
      ],
      totalCount: 1,
      limit: 20,
      offset: 0,
    });

    render(<VideoList />);

    const thumbnail = await screen.findByRole("img", { name: "Demo thumbnail" });
    expect(Number.parseFloat(thumbnail.style.backgroundPosition)).toBeCloseTo(100 / 3);
    expect(screen.getByRole("link", { name: "Open Demo" })).toHaveAttribute("href", "/videos/video-123");
    expect(screen.queryByRole("link", { name: "View details" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Delete" })).not.toBeInTheDocument();
  });
});
