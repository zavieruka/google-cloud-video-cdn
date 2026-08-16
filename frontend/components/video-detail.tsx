"use client";

import Link from "next/link";
import { useEffect, useState } from "react";

import { ApiError, getVideo, selectThumbnail } from "../lib/api";
import type { Video, VideoStatus as VideoStatusValue } from "../lib/types";
import { HlsPlayer } from "./hls-player";
import { Thumbnail } from "./thumbnail";
import { VideoStatus, videoStatusLabel } from "./video-status";

const pollableStatuses = new Set<VideoStatusValue>(["pending", "uploaded", "processing"]);
const pollingIntervalMs = 3000;

function errorMessage(error: unknown): string {
  if (error instanceof ApiError || error instanceof Error) {
    return error.message;
  }
  return "Unable to load this video.";
}

export function VideoDetail({ videoId }: { videoId: string }) {
  const [video, setVideo] = useState<Video>();
  const [error, setError] = useState<string>();
  const [selectingIndex, setSelectingIndex] = useState<number>();
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
      setThumbnailError(undefined);
    } catch (selectionError) {
      setThumbnailError(errorMessage(selectionError));
    } finally {
      setSelectingIndex(undefined);
    }
  };

  if (error) {
    return (
      <section className="rounded-xl border border-rose-200 bg-rose-50 p-6">
        <h1 className="text-lg font-semibold text-rose-950">Could not load video</h1>
        <p className="mt-2 text-sm text-rose-800">{error}</p>
        <Link className="mt-4 inline-block text-sm font-semibold text-rose-900 underline" href="/">
          Back to videos
        </Link>
      </section>
    );
  }

  if (!video) {
    return <p className="text-sm text-slate-600">Loading video…</p>;
  }

  const thumbnail = video.thumbnail;

  return (
    <article className="space-y-6">
      <Link className="text-sm font-semibold text-blue-700 underline" href="/">
        Back to videos
      </Link>
      <header className="space-y-3">
        <VideoStatus status={video.status} />
        <h1 className="text-3xl font-bold tracking-tight text-slate-950">{video.title}</h1>
        {video.description ? <p className="max-w-2xl text-slate-600">{video.description}</p> : null}
        <p className="text-sm text-slate-500">{video.fileName}</p>
      </header>

      {pollableStatuses.has(video.status) ? (
        <p className="rounded-lg bg-amber-50 p-4 text-sm text-amber-900" role="status">
          {videoStatusLabel(video.status)}. This page checks again every three seconds.
        </p>
      ) : null}

      {video.status === "failed" ? (
        <section className="rounded-xl border border-rose-200 bg-rose-50 p-5">
          <h2 className="font-semibold text-rose-950">Processing failed</h2>
          <p className="mt-2 text-sm text-rose-800">{video.lastError ?? "The video could not be processed."}</p>
          <Link className="mt-4 inline-block text-sm font-semibold text-rose-900 underline" href="/upload">
            Upload another video
          </Link>
        </section>
      ) : null}

      {video.status === "ready" && video.manifestUrl ? <HlsPlayer manifestUrl={video.manifestUrl} /> : null}
      {video.status === "ready" && !video.manifestUrl ? (
        <p className="rounded-lg bg-rose-50 p-4 text-sm text-rose-800">
          The video is ready, but its playback manifest is unavailable.
        </p>
      ) : null}

      {video.status === "ready" && thumbnail ? (
        <section className="space-y-3">
          <h2 className="text-lg font-semibold text-slate-950">Thumbnail</h2>
          <div className="max-w-2xl">
            <Thumbnail alt={`${video.title} thumbnail`} index={thumbnail.selectedIndex} url={thumbnail.url} />
          </div>
          <div aria-label="Thumbnail candidates" className="grid grid-cols-3 gap-2 sm:grid-cols-4" role="group">
            {Array.from({ length: 12 }, (_, index) => (
              <button
                aria-label={`Select thumbnail candidate ${index + 1}`}
                aria-pressed={thumbnail.selectedIndex === index}
                className="rounded-md focus:outline-none focus:ring-2 focus:ring-blue-600 disabled:opacity-50"
                disabled={selectingIndex !== undefined}
                key={index}
                onClick={() => void selectCandidate(index)}
                type="button"
              >
                <Thumbnail alt={`Thumbnail candidate ${index + 1}`} index={index} url={thumbnail.url} />
              </button>
            ))}
          </div>
          {thumbnailError ? <p className="text-sm text-rose-800">{thumbnailError}</p> : null}
        </section>
      ) : null}
    </article>
  );
}
