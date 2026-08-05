import { afterEach, describe, expect, it, vi } from "vitest";

import { uploadFile } from "./api";

class FakeXMLHttpRequest {
  static instances: FakeXMLHttpRequest[] = [];

  status = 200;
  upload: XMLHttpRequestUpload = {} as XMLHttpRequestUpload;
  onload: (() => void) | null = null;
  onerror: (() => void) | null = null;
  onabort: (() => void) | null = null;
  open = vi.fn();
  setRequestHeader = vi.fn();
  send = vi.fn((body: Document | XMLHttpRequestBodyInit | null) => {
    const onProgress = this.upload.onprogress;
    onProgress?.call(this as unknown as XMLHttpRequest, {
      lengthComputable: true,
      loaded: 25,
      total: 100,
    } as ProgressEvent<XMLHttpRequestEventTarget>);
    this.onload?.();
    return body;
  });

  constructor() {
    FakeXMLHttpRequest.instances.push(this);
  }
}

describe("uploadFile", () => {
  afterEach(() => {
    FakeXMLHttpRequest.instances = [];
    vi.unstubAllGlobals();
  });

  it("uses a direct signed PUT with the selected file's exact MIME type", async () => {
    vi.stubGlobal("XMLHttpRequest", FakeXMLHttpRequest);
    const file = new File(["video bytes"], "demo.mp4", { type: "video/mp4" });
    const onProgress = vi.fn();

    await uploadFile("https://storage.googleapis.com/signed-upload", file, onProgress);

    const request = FakeXMLHttpRequest.instances[0];
    expect(request.open).toHaveBeenCalledWith(
      "PUT",
      "https://storage.googleapis.com/signed-upload",
      true,
    );
    expect(request.setRequestHeader).toHaveBeenCalledWith("Content-Type", "video/mp4");
    expect(request.send).toHaveBeenCalledWith(file);
    expect(onProgress).toHaveBeenCalledWith(0.25);
  });
});
