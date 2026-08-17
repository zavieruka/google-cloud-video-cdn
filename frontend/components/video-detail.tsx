"use client";

import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";

import {
  ApiError,
  confirmThumbnailUpload,
  deleteVideo,
  getVideo,
  requestThumbnailUploadUrl,
  selectThumbnail,
  uploadFile,
} from "../lib/api";
import type { Video, VideoStatus as VideoStatusValue } from "../lib/types";
import { HlsPlayer } from "./hls-player";
import { Thumbnail } from "./thumbnail";
import { VideoStatus, videoStatusLabel } from "./video-status";

const pollableStatuses = new Set<VideoStatusValue>(["pending", "uploaded", "processing"]);
const pollingIntervalMs = 3000;
const thumbnailCandidateCount = 12;
const thumbnailCandidatePageSize = 4;

function errorMessage(error: unknown): string {
  if (error instanceof ApiError || error instanceof Error) {
    return error.message;
  }
  return "Unable to load this video.";
}

export function VideoDetail({ videoId }: { videoId: string }) {
  const router = useRouter();
  const [video, setVideo] = useState<Video>();
  const [error, setError] = useState<string>();
  const [deleteError, setDeleteError] = useState<string>();
  const [isDeleting, setIsDeleting] = useState(false);
  const [selectingIndex, setSelectingIndex] = useState<number>();
  const [thumbnailRevision, setThumbnailRevision] = useState(0);
  const [thumbnailCandidateStart, setThumbnailCandidateStart] = useState(0);
  const [thumbnailFile, setThumbnailFile] = useState<File>();
  const [uploadingThumbnail, setUploadingThumbnail] = useState(false);
  const [thumbnailError, setThumbnailError] = useState<string>();

  useEffect(() => {
    let disposed = false;
    let timer: number | undefined;
    const controller = new AbortController();

    const load = async () => {
      try {
        const nextVideo = await getVideo(videoId, controller.signal);
        if (disposed) {
          return;
        }

        setVideo(nextVideo);
        setError(undefined);
        if (pollableStatuses.has(nextVideo.status)) {
          timer = window.setTimeout(load, pollingIntervalMs);
        }
      } catch (loadError) {
        if (!disposed && !(loadError instanceof DOMException && loadError.name === "AbortError")) {
          setError(errorMessage(loadError));
        }
      }
    };

    void load();

    return () => {
      disposed = true;
      controller.abort();
      if (timer !== undefined) {
        window.clearTimeout(timer);
      }
    };
  }, [videoId]);

  const selectCandidate = async (selectedIndex: number) => {
    if (!video?.thumbnail) {
      return;
    }

    setSelectingIndex(selectedIndex);
    try {
      const updatedVideo = await selectThumbnail(video.id, selectedIndex);
      setVideo(updatedVideo);
      setThumbnailRevision((revision) => revision + 1);
      setThumbnailError(undefined);
    } catch (selectionError) {
      setThumbnailError(errorMessage(selectionError));
    } finally {
      setSelectingIndex(undefined);
    }
  };

  const uploadCustomThumbnail = async () => {
    if (!video || !thumbnailFile) {
      return;
    }

    setUploadingThumbnail(true);
    try {
      const { uploadUrl } = await requestThumbnailUploadUrl(video.id, {
        mimeType: thumbnailFile.type,
        fileSize: thumbnailFile.size,
      });
      await uploadFile(uploadUrl, thumbnailFile);
      const updatedVideo = await confirmThumbnailUpload(video.id);
      setVideo(updatedVideo);
      setThumbnailRevision((revision) => revision + 1);
      setThumbnailFile(undefined);
      setThumbnailError(undefined);
    } catch (uploadError) {
      setThumbnailError(errorMessage(uploadError));
    } finally {
      setUploadingThumbnail(false);
    }
  };

  const removeVideo = async () => {
    if (!video) {
      return;
    }

    setIsDeleting(true);
    setDeleteError(undefined);
    try {
      await deleteVideo(video.id);
      router.push("/");
    } catch (deleteError) {
      setDeleteError(errorMessage(deleteError));
    } finally {
      setIsDeleting(false);
    }
  };

  if (error) {
    return (
      <section className="mx-auto max-w-6xl rounded-2xl border border-rose-200 bg-rose-50 p-6">
        <h1 className="text-lg font-semibold text-rose-950">Could not load video</h1>
        <p className="mt-2 text-sm text-rose-800">{error}</p>
      </section>
    );
  }

  if (!video) {
    return <p className="mx-auto max-w-6xl rounded-2xl border border-slate-200 bg-white p-6 text-sm text-slate-600 shadow-sm">Loading video…</p>;
  }

  const thumbnail = video.thumbnail;
  const thumbnailURL = thumbnail ? `${thumbnail.url}?thumbnailRevision=${thumbnailRevision}` : "";
  const candidatesURL = thumbnail ? `${thumbnail.candidatesUrl}?thumbnailRevision=${thumbnailRevision}` : "";
  const visibleCandidateIndexes = Array.from(
    { length: Math.min(thumbnailCandidatePageSize, thumbnailCandidateCount - thumbnailCandidateStart) },
    (_, offset) => thumbnailCandidateStart + offset,
  );
  const canShowPreviousCandidates = thumbnailCandidateStart > 0;
  const canShowNextCandidates = thumbnailCandidateStart + thumbnailCandidatePageSize < thumbnailCandidateCount;

  return (
    <article className="mx-auto max-w-6xl space-y-8">
      <header className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm sm:p-8">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div>
            <p className="text-sm font-semibold uppercase tracking-wider text-blue-700">Video details</p>
            <h1 className="mt-2 text-3xl font-bold tracking-tight text-slate-950 sm:text-4xl">{video.title}</h1>
            {video.description ? <p className="mt-3 max-w-3xl text-slate-600">{video.description}</p> : null}
            <p className="mt-4 text-sm text-slate-500">{video.fileName}</p>
          </div>
          <div className="flex items-center gap-3">
            <VideoStatus status={video.status} />
            <button
              className="rounded-lg border border-rose-200 px-3 py-2 text-sm font-semibold text-rose-700 transition hover:bg-rose-50 disabled:cursor-not-allowed disabled:opacity-50"
              disabled={isDeleting}
              onClick={() => void removeVideo()}
              type="button"
            >
              {isDeleting ? "Deleting…" : "Delete video"}
            </button>
          </div>
        </div>
        {deleteError ? <p className="mt-4 text-sm text-rose-800">{deleteError}</p> : null}
      </header>

      {pollableStatuses.has(video.status) ? (
        <p className="rounded-xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-900" role="status">
          {videoStatusLabel(video.status)}. This page checks again every three seconds.
        </p>
      ) : null}

      {video.status === "failed" ? (
        <section className="rounded-2xl border border-rose-200 bg-rose-50 p-6">
          <h2 className="font-semibold text-rose-950">Processing failed</h2>
          <p className="mt-2 text-sm text-rose-800">{video.lastError ?? "The video could not be processed."}</p>
        </section>
      ) : null}

      {video.status === "ready" && video.manifestUrl ? (
        <section className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm sm:p-6">
          <div className="mb-4">
            <h2 className="text-lg font-semibold text-slate-950">Playback</h2>
            <p className="mt-1 text-sm text-slate-500">Adaptive HLS delivery from the processed video output.</p>
          </div>
          <HlsPlayer manifestUrl={video.manifestUrl} />
        </section>
      ) : null}
      {video.status === "ready" && !video.manifestUrl ? (
        <p className="rounded-xl border border-rose-200 bg-rose-50 p-4 text-sm text-rose-800">
          The video is ready, but its playback manifest is unavailable.
        </p>
      ) : null}

      {video.status === "ready" && thumbnail ? (
        <section className="space-y-5 rounded-2xl border border-slate-200 bg-white p-5 shadow-sm sm:p-6">
          <div>
            <h2 className="text-lg font-semibold text-slate-950">Thumbnail</h2>
            <p className="mt-1 text-sm text-slate-500">Choose a generated frame or upload your own image.</p>
          </div>
          <div className="grid gap-4 lg:grid-cols-2">
            <div className="rounded-xl border border-slate-200 bg-slate-50 p-4">
              <h3 className="text-sm font-semibold text-slate-900">Current thumbnail</h3>
              <div className="mt-3 max-w-sm overflow-hidden rounded-lg border border-slate-100 bg-white">
                <Thumbnail alt={`${video.title} thumbnail`} index={thumbnail.selectedIndex} url={thumbnailURL} />
              </div>
            </div>
            <div className="rounded-xl border border-slate-200 bg-slate-50 p-4">
              <label className="block text-sm font-semibold text-slate-900" htmlFor="custom-thumbnail">
                Custom thumbnail image
              </label>
              <input
                accept="image/jpeg,image/png,image/webp"
                className="mt-3 block w-full text-sm text-slate-700 file:mr-4 file:rounded-md file:border-0 file:bg-blue-100 file:px-3 file:py-1.5 file:font-semibold file:text-blue-700"
                id="custom-thumbnail"
                onChange={(event) => setThumbnailFile(event.target.files?.[0])}
                type="file"
              />
              <p className="mt-3 text-sm text-slate-600">
                JPEG, PNG, or WebP up to 10 MB. Use at least 1280×720 for best results; the original image is stored without resizing.
              </p>
              <button
                className="mt-4 rounded-lg bg-blue-600 px-3 py-2 text-sm font-semibold text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
                disabled={!thumbnailFile || uploadingThumbnail}
                onClick={() => void uploadCustomThumbnail()}
                type="button"
              >
                {uploadingThumbnail ? "Uploading thumbnail…" : "Upload custom thumbnail"}
              </button>
            </div>
          </div>
          <div aria-label="Thumbnail candidates" className="rounded-xl border border-slate-200 p-4" role="group">
            <div className="mb-3 flex items-center justify-between gap-3">
              <div>
                <h3 className="text-sm font-semibold text-slate-900">Suggested frames</h3>
                <p aria-live="polite" className="mt-1 text-sm text-slate-500">
                  Showing {thumbnailCandidateStart + 1}–{thumbnailCandidateStart + visibleCandidateIndexes.length} of {thumbnailCandidateCount}
                </p>
              </div>
              <div className="flex gap-2">
                <button
                  aria-label="Show previous thumbnail suggestions"
                  className="grid h-9 w-9 place-items-center rounded-lg border border-slate-200 text-slate-600 transition hover:border-blue-200 hover:bg-blue-50 hover:text-blue-700 disabled:cursor-not-allowed disabled:opacity-40"
                  disabled={!canShowPreviousCandidates || selectingIndex !== undefined}
                  onClick={() => setThumbnailCandidateStart((start) => Math.max(0, start - thumbnailCandidatePageSize))}
                  type="button"
                >
                  <svg aria-hidden="true" className="h-4 w-4" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                    <path d="m15 18-6-6 6-6" />
                  </svg>
                </button>
                <button
                  aria-label="Show next thumbnail suggestions"
                  className="grid h-9 w-9 place-items-center rounded-lg border border-slate-200 text-slate-600 transition hover:border-blue-200 hover:bg-blue-50 hover:text-blue-700 disabled:cursor-not-allowed disabled:opacity-40"
                  disabled={!canShowNextCandidates || selectingIndex !== undefined}
                  onClick={() => setThumbnailCandidateStart((start) => Math.min(thumbnailCandidateCount - thumbnailCandidatePageSize, start + thumbnailCandidatePageSize))}
                  type="button"
                >
                  <svg aria-hidden="true" className="h-4 w-4" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                    <path d="m9 18 6-6-6-6" />
                  </svg>
                </button>
              </div>
            </div>
            <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
              {visibleCandidateIndexes.map((index) => (
              <button
                aria-label={`Select thumbnail candidate ${index + 1}`}
                aria-pressed={thumbnail.selectedIndex === index}
                className="rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-600 disabled:opacity-50"
                disabled={selectingIndex !== undefined}
                key={index}
                onClick={() => void selectCandidate(index)}
                type="button"
              >
                <Thumbnail alt={`Thumbnail candidate ${index + 1}`} index={index} url={candidatesURL} />
              </button>
              ))}
            </div>
          </div>
          {thumbnailError ? <p className="text-sm text-rose-800">{thumbnailError}</p> : null}
        </section>
      ) : null}
    </article>
  );
}
