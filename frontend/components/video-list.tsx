"use client";

import Link from "next/link";
import { useEffect, useState } from "react";

import { ApiError, listVideos } from "../lib/api";
import type { Video } from "../lib/types";
import { Thumbnail } from "./thumbnail";
import { VideoStatus } from "./video-status";

function errorMessage(error: unknown): string {
  if (error instanceof ApiError || error instanceof Error) {
    return error.message;
  }
  return "Unable to load videos.";
}

export function VideoList() {
  const [videos, setVideos] = useState<Video[]>();
  const [error, setError] = useState<string>();

  useEffect(() => {
    let isMounted = true;

    void listVideos()
      .then((response) => {
        if (isMounted) {
          setVideos(response.videos);
          setError(undefined);
        }
      })
      .catch((loadError: unknown) => {
        if (isMounted) {
          setError(errorMessage(loadError));
        }
      });

    return () => {
      isMounted = false;
    };
  }, []);

  return (
    <section className="mx-auto max-w-6xl space-y-8">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <p className="text-sm font-semibold uppercase tracking-wider text-blue-700">Media library</p>
          <h1 className="mt-2 text-3xl font-bold tracking-tight text-slate-950 sm:text-4xl">Videos</h1>
          <p className="mt-3 text-slate-600">Upload, process, and play HLS video from one place.</p>
        </div>
        <Link className="rounded-lg bg-blue-600 px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition hover:bg-blue-700" href="/upload">
          Upload video
        </Link>
      </div>

      {error ? <p className="rounded-xl border border-rose-200 bg-rose-50 p-4 text-sm text-rose-800">{error}</p> : null}
      {!videos && !error ? <p className="rounded-xl border border-slate-200 bg-white p-5 text-sm text-slate-600 shadow-sm">Loading videos…</p> : null}
      {videos?.length === 0 ? (
        <div className="rounded-2xl border border-dashed border-slate-300 bg-white p-10 text-center shadow-sm">
          <h2 className="font-semibold text-slate-900">Your library is empty</h2>
          <p className="mx-auto mt-2 max-w-md text-sm text-slate-600">Upload a video to see the processing pipeline in action.</p>
          <Link className="mt-5 inline-flex rounded-lg bg-blue-600 px-4 py-2.5 text-sm font-semibold text-white hover:bg-blue-700" href="/upload">
            Upload your first video
          </Link>
        </div>
      ) : null}
      {videos?.length ? (
        <ul className="grid gap-5 md:grid-cols-2 xl:grid-cols-3">
          {videos.map((video) => (
            <li className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm transition hover:-translate-y-0.5 hover:shadow-md" key={video.id}>
              <Link aria-label={`Open ${video.title}`} className="block h-full focus:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-blue-600" href={`/videos/${video.id}`}>
                {video.status === "ready" && video.thumbnail ? (
                  <Thumbnail alt={`${video.title} thumbnail`} index={video.thumbnail.selectedIndex} url={video.thumbnail.url} />
                ) : null}
                <div className="p-5">
                  <div className="flex items-start justify-between gap-3">
                    <div>
                      <p className="font-semibold text-slate-950">{video.title}</p>
                      <p className="mt-1 text-sm text-slate-600">{video.fileName}</p>
                    </div>
                    <VideoStatus status={video.status} />
                  </div>
                </div>
              </Link>
            </li>
          ))}
        </ul>
      ) : null}
    </section>
  );
}
