import { resolveApiUrl } from "../lib/api";

const columns = 4;
const rows = 3;

export function Thumbnail({ alt, index, url }: { alt: string; index?: number; url: string }) {
	if (index === undefined) {
		// The API redirects to a short-lived signed object URL; do not proxy it through Next's optimizer.
		// eslint-disable-next-line @next/next/no-img-element
		return <img alt={alt} className="aspect-video w-full rounded-md bg-slate-200 object-cover" src={resolveApiUrl(url)} />;
	}

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
