"use client";

import { useRouter } from "next/navigation";

import { UploadForm } from "../../components/upload-form";

export default function UploadPage() {
  const router = useRouter();

  return (
    <section className="mx-auto max-w-3xl space-y-8">
      <div>
        <p className="text-sm font-semibold uppercase tracking-wider text-blue-700">New upload</p>
        <h1 className="mt-2 text-3xl font-bold tracking-tight text-slate-950 sm:text-4xl">Upload a video</h1>
        <p className="mt-3 max-w-2xl text-slate-600">Send a source file directly to private Cloud Storage, then follow its processing progress from the video library.</p>
      </div>
      <div className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm sm:p-8">
        <h2 className="text-lg font-semibold text-slate-950">Video details</h2>
        <p className="mt-1 text-sm text-slate-500">A title and a supported video file are required.</p>
        <div className="mt-6">
          <UploadForm onCompleted={(videoID) => router.push(`/videos/${videoID}`)} />
        </div>
      </div>
    </section>
  );
}
