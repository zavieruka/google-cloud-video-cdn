"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";

import { UploadForm } from "../../components/upload-form";

export default function UploadPage() {
  const router = useRouter();

  return (
    <section className="mx-auto max-w-2xl space-y-6">
      <Link className="text-sm font-semibold text-blue-700 underline" href="/">
        Back to videos
      </Link>
      <div>
        <h1 className="text-3xl font-bold tracking-tight text-slate-950">Upload a video</h1>
        <p className="mt-2 text-slate-600">The video uploads directly to private Cloud Storage.</p>
      </div>
      <UploadForm onCompleted={(videoID) => router.push(`/videos/${videoID}`)} />
    </section>
  );
}
