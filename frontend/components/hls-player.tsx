"use client";

import Hls from "hls.js";
import { useEffect, useRef, useState } from "react";

import { resolveApiUrl } from "../lib/api";

export function HlsPlayer({ manifestUrl }: { manifestUrl: string }) {
  return <HlsPlayerSource key={manifestUrl} manifestUrl={manifestUrl} />;
}

function HlsPlayerSource({ manifestUrl }: { manifestUrl: string }) {
  const videoRef = useRef<HTMLVideoElement>(null);
  const [error, setError] = useState<string>();

  useEffect(() => {
    const video = videoRef.current;
    if (!video) {
      return;
    }

    const source = resolveApiUrl(manifestUrl);
    const onNativeError = () => setError("The video could not be played. Please try again.");
    video.addEventListener("error", onNativeError);

    if (video.canPlayType("application/vnd.apple.mpegurl")) {
      video.src = source;
      return () => {
        video.removeEventListener("error", onNativeError);
        video.removeAttribute("src");
      };
    }

    if (!Hls.isSupported()) {
      let isMounted = true;
      queueMicrotask(() => {
        if (isMounted) {
          setError("HLS playback is not supported by this browser.");
        }
      });
      return () => {
        isMounted = false;
        video.removeEventListener("error", onNativeError);
      };
    }

    const hls = new Hls();
    hls.on(Hls.Events.ERROR, (_event, data) => {
      if (data.fatal) {
        setError("The video could not be played. Please try again.");
      }
    });
    hls.loadSource(source);
    hls.attachMedia(video);

    return () => {
      video.removeEventListener("error", onNativeError);
      hls.destroy();
    };
  }, [manifestUrl]);

  return (
    <div className="space-y-3">
      <video
        ref={videoRef}
        aria-label="Video player"
        className="aspect-video w-full rounded-xl bg-slate-950"
        controls
      />
      {error ? <p className="text-sm text-rose-700">{error}</p> : null}
    </div>
  );
}
