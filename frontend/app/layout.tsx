import type { Metadata } from "next";
import type { ReactNode } from "react";

import "./globals.css";

export const metadata: Metadata = {
  title: "Video Platform",
  description: "Example web consumer for the Video Platform API",
};

export default function RootLayout({ children }: Readonly<{ children: ReactNode }>) {
  return (
    <html lang="en">
      <body>
        <main className="mx-auto min-h-screen max-w-5xl px-5 py-10 sm:px-8">{children}</main>
      </body>
    </html>
  );
}
