import { afterEach, describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

const api = vi.hoisted(() => ({
  requestUploadUrl: vi.fn(),
  uploadFile: vi.fn(),
  confirmUpload: vi.fn(),
  failUpload: vi.fn(),
}));

vi.mock("../lib/api", () => api);

import { UploadForm } from "./upload-form";

describe("UploadForm", () => {
  afterEach(() => {
    vi.clearAllMocks();
  });

  it("requests, uploads, then confirms a selected video", async () => {
    const file = new File(["video bytes"], "demo.mp4", { type: "video/mp4" });
    const onCompleted = vi.fn();
    api.requestUploadUrl.mockResolvedValue({
      videoId: "video-123",
      uploadUrl: "https://storage.googleapis.com/signed-upload",
    });
    api.uploadFile.mockResolvedValue(undefined);
    api.confirmUpload.mockResolvedValue({ id: "video-123", status: "uploaded" });

    render(<UploadForm onCompleted={onCompleted} />);
    const user = userEvent.setup();

    await user.type(screen.getByLabelText("Title"), "Demo");
    await user.type(screen.getByLabelText("Description"), "A short test video");
    fireEvent.change(screen.getByLabelText("Video file"), { target: { files: [file] } });
    await user.click(screen.getByRole("button", { name: "Upload video" }));

    await waitFor(() => expect(onCompleted).toHaveBeenCalledWith("video-123"));

    expect(api.requestUploadUrl).toHaveBeenCalledWith({
      title: "Demo",
      description: "A short test video",
      fileName: "demo.mp4",
      fileSize: file.size,
      mimeType: "video/mp4",
    });
    expect(api.uploadFile).toHaveBeenCalledWith(
      "https://storage.googleapis.com/signed-upload",
      file,
      expect.any(Function),
    );
    expect(api.confirmUpload).toHaveBeenCalledWith("video-123", expect.any(String));
  });
});
