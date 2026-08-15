"use client";

import Link from "next/link";
import { useEffect, useState } from "react";

import { ApiError, deleteVideo, listVideos } from "../lib/api";
import type { Video } from "../lib/types";
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
  const [deletingID, setDeletingID] = useState<string>();

  const loadVideos = async () => {
    try {
      const response = await listVideos();
      setVideos(response.videos);
      setError(undefined);
    } catch (loadError) {
      setError(errorMessage(loadError));
    }
  };

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

  const removeVideo = async (videoID: string) => {
    setDeletingID(videoID);
    try {
      await deleteVideo(videoID);
      await loadVideos();
    } catch (deleteError) {
      setError(errorMessage(deleteError));
    } finally {
      setDeletingID(undefined);
    }
  };

  return (
    <section className="space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-slate-950">Videos</h1>
          <p className="mt-1 text-slate-600">Upload, process, and play HLS video.</p>
        </div>
        <Link className="rounded-md bg-blue-700 px-4 py-2 font-semibold text-white" href="/upload">
          Upload video
        </Link>
      </div>

      {error ? <p className="rounded-md bg-rose-50 p-3 text-sm text-rose-800">{error}</p> : null}
      {!videos && !error ? <p className="text-sm text-slate-600">Loading videos…</p> : null}
      {videos?.length === 0 ? (
        <div className="rounded-xl border border-dashed border-slate-300 p-8 text-center">
          <h2 className="font-semibold text-slate-900">No videos yet</h2>
          <p className="mt-2 text-sm text-slate-600">Upload a video to see the processing pipeline in action.</p>
        </div>
      ) : null}
      {videos?.length ? (
        <ul className="grid gap-4 sm:grid-cols-2">
          {videos.map((video) => (
            <li className="rounded-xl border border-slate-200 p-5 shadow-sm" key={video.id}>
              <div className="flex items-start justify-between gap-3">
                <div>
                  <Link className="font-semibold text-slate-950 hover:underline" href={`/videos/${video.id}`}>
                    {video.title}
                  </Link>
                  <p className="mt-1 text-sm text-slate-600">{video.fileName}</p>
                </div>
                <VideoStatus status={video.status} />
              </div>
              <div className="mt-5 flex items-center justify-between gap-3">
                <Link className="text-sm font-semibold text-blue-700 underline" href={`/videos/${video.id}`}>
                  View details
                </Link>
                <button
                  className="text-sm font-semibold text-rose-700 underline disabled:text-slate-400"
                  disabled={deletingID === video.id}
                  onClick={() => void removeVideo(video.id)}
                  type="button"
                >
                  {deletingID === video.id ? "Deleting…" : "Delete"}
                </button>
              </div>
            </li>
          ))}
        </ul>
      ) : null}
    </section>
  );
}
