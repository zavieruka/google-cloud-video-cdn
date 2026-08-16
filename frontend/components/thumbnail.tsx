import { resolveApiUrl } from "../lib/api";

const columns = 4;
const rows = 3;

export function Thumbnail({ alt, index, url }: { alt: string; index: number; url: string }) {
  const column = index % columns;
  const row = Math.floor(index / columns);
  const horizontalPosition = (column / (columns - 1)) * 100;
  const verticalPosition = (row / (rows - 1)) * 100;

  return (
    <div
      aria-label={alt}
      className="aspect-video rounded-md bg-slate-200 bg-cover"
      role="img"
      style={{
        backgroundImage: `url("${resolveApiUrl(url)}")`,
        backgroundPosition: `${horizontalPosition}% ${verticalPosition}%`,
        backgroundSize: `${columns * 100}% ${rows * 100}%`,
      }}
    />
  );
}
