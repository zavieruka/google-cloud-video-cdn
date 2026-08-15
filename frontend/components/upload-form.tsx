"use client";

import { FormEvent, useState } from "react";

import { ApiError, confirmUpload, failUpload, requestUploadUrl, uploadFile } from "../lib/api";

const maxUploadSizeBytes = 500 * 1024 * 1024;
const allowedMimeTypes = new Set(["video/mp4", "video/quicktime", "video/x-msvideo", "video/x-matroska"]);

function messageFor(error: unknown): string {
  if (error instanceof ApiError || error instanceof Error) {
    return error.message;
  }
  return "The upload could not be completed.";
}

export function UploadForm({ onCompleted }: { onCompleted: (videoId: string) => void }) {
  const [file, setFile] = useState<File>();
  const [progress, setProgress] = useState<number>();
  const [error, setError] = useState<string>();
  const [isSubmitting, setIsSubmitting] = useState(false);

  const onSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const title = String(form.get("title") ?? "").trim();
    const description = String(form.get("description") ?? "").trim();

    if (!title) {
      setError("A title is required.");
      return;
    }
    if (!file) {
      setError("Choose a video file to upload.");
      return;
    }
    if (!allowedMimeTypes.has(file.type)) {
      setError("Choose an MP4, MOV, AVI, or MKV video file.");
      return;
    }
    if (file.size > maxUploadSizeBytes) {
      setError("The selected video exceeds the 500 MB upload limit.");
      return;
    }

    setError(undefined);
    setProgress(0);
    setIsSubmitting(true);

    try {
      const signedUpload = await requestUploadUrl({
        title,
        description,
        fileName: file.name,
        fileSize: file.size,
        mimeType: file.type,
      });

      try {
        await uploadFile(signedUpload.uploadUrl, file, setProgress);
      } catch (uploadError) {
        // The signed PUT may have failed after the backend created its pending
        // record. Record the failure best-effort, but preserve the original
        // browser/storage error for the user.
        try {
          await failUpload(signedUpload.videoId, {
            error: "upload_failed",
            message: messageFor(uploadError),
          });
        } catch {
          // The original upload error remains the useful one to surface.
        }
        throw uploadError;
      }

      await confirmUpload(signedUpload.videoId, new Date().toISOString());
      onCompleted(signedUpload.videoId);
    } catch (submitError) {
      setError(messageFor(submitError));
      setProgress(undefined);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <form className="space-y-5" onSubmit={onSubmit}>
      <label className="block space-y-1.5 text-sm font-medium text-slate-800" htmlFor="title">
        Title
        <input
          className="block w-full rounded-md border border-slate-300 px-3 py-2 text-base"
          id="title"
          maxLength={200}
          name="title"
          required
          type="text"
        />
      </label>

      <label className="block space-y-1.5 text-sm font-medium text-slate-800" htmlFor="description">
        Description
        <textarea
          className="block w-full rounded-md border border-slate-300 px-3 py-2 text-base"
          id="description"
          name="description"
          rows={4}
        />
      </label>

      <label className="block space-y-1.5 text-sm font-medium text-slate-800" htmlFor="video-file">
        Video file
        <input
          accept="video/mp4,video/quicktime,video/x-msvideo,video/x-matroska"
          className="block w-full text-sm"
          id="video-file"
          name="video-file"
          onChange={(event) => setFile(event.target.files?.[0])}
          type="file"
        />
      </label>

      {progress !== undefined ? (
        <div aria-live="polite" className="space-y-1">
          <progress className="w-full" max="1" value={progress} />
          <p className="text-sm text-slate-600">Uploading {Math.round(progress * 100)}%</p>
        </div>
      ) : null}
      {error ? <p className="rounded-md bg-rose-50 p-3 text-sm text-rose-800">{error}</p> : null}

      <button
        className="rounded-md bg-blue-700 px-4 py-2 font-semibold text-white disabled:cursor-not-allowed disabled:bg-slate-400"
        disabled={isSubmitting}
        type="submit"
      >
        {isSubmitting ? "Uploading…" : "Upload video"}
      </button>
    </form>
  );
}
