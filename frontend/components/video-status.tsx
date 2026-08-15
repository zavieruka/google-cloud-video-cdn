import type { VideoStatus as VideoStatusValue } from "../lib/types";

const labels: Record<VideoStatusValue, string> = {
  pending: "Pending upload",
  uploaded: "Uploaded",
  processing: "Processing",
  ready: "Ready",
  failed: "Failed",
};

const colors: Record<VideoStatusValue, string> = {
  pending: "bg-slate-100 text-slate-700",
  uploaded: "bg-blue-100 text-blue-800",
  processing: "bg-amber-100 text-amber-800",
  ready: "bg-emerald-100 text-emerald-800",
  failed: "bg-rose-100 text-rose-800",
};

export function VideoStatus({ status }: { status: VideoStatusValue }) {
  return (
    <span className={`inline-flex rounded-full px-2.5 py-1 text-xs font-semibold ${colors[status]}`}>
      {labels[status]}
    </span>
  );
}

export function videoStatusLabel(status: VideoStatusValue): string {
  return labels[status];
}
