"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

export function AppSidebar() {
  const pathname = usePathname();
  const videosActive = pathname === "/" || pathname.startsWith("/videos/");
  const uploadActive = pathname === "/upload";

  return (
    <aside className="border-b border-slate-200 bg-white lg:sticky lg:top-0 lg:flex lg:h-screen lg:w-64 lg:shrink-0 lg:flex-col lg:border-r lg:border-b-0">
      <div className="flex h-16 items-center border-b border-slate-100 px-4 lg:h-20 lg:px-6">
        <Link className="flex items-center gap-3" href="/">
          <span className="grid h-9 w-9 place-items-center rounded-xl bg-blue-600 text-sm font-bold text-white shadow-sm">VP</span>
          <span>
            <span className="block text-sm font-bold tracking-tight text-slate-950">Video Platform</span>
            <span className="block text-xs text-slate-500">Media workspace</span>
          </span>
        </Link>
      </div>

      <nav aria-label="Main navigation" className="overflow-x-auto px-3 py-3 lg:px-4 lg:py-6">
        <p className="mb-2 hidden px-3 text-xs font-semibold uppercase tracking-wider text-slate-400 lg:block">Workspace</p>
        <ul className="flex min-w-max gap-1 lg:block lg:space-y-1">
          <li>
            <Link
              aria-current={videosActive ? "page" : undefined}
              className={`flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-semibold transition ${
                videosActive ? "bg-blue-50 text-blue-700" : "text-slate-600 hover:bg-slate-100 hover:text-slate-950"
              }`}
              href="/"
            >
              <svg aria-hidden="true" className="h-5 w-5" fill="none" stroke="currentColor" strokeWidth="1.8" viewBox="0 0 24 24">
                <path d="M4 6.75A2.75 2.75 0 0 1 6.75 4h10.5A2.75 2.75 0 0 1 20 6.75v10.5A2.75 2.75 0 0 1 17.25 20H6.75A2.75 2.75 0 0 1 4 17.25V6.75Z" />
                <path d="m10 9 5 3-5 3V9Z" />
              </svg>
              Videos
            </Link>
          </li>
          <li>
            <Link
              aria-current={uploadActive ? "page" : undefined}
              className={`flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-semibold transition ${
                uploadActive ? "bg-blue-50 text-blue-700" : "text-slate-600 hover:bg-slate-100 hover:text-slate-950"
              }`}
              href="/upload"
            >
              <svg aria-hidden="true" className="h-5 w-5" fill="none" stroke="currentColor" strokeWidth="1.8" viewBox="0 0 24 24">
                <path d="M12 16V4m0 0L7.5 8.5M12 4l4.5 4.5M5 15.5V19a1 1 0 0 0 1 1h12a1 1 0 0 0 1-1v-3.5" />
              </svg>
              Upload
            </Link>
          </li>
        </ul>
      </nav>

      <div className="mt-auto hidden border-t border-slate-100 px-6 py-5 lg:block">
        <p className="text-xs leading-5 text-slate-500">Upload a source video, follow its processing state, and watch the HLS output here.</p>
      </div>
    </aside>
  );
}
