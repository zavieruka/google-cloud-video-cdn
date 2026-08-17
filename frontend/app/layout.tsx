import type { Metadata } from "next";
import type { ReactNode } from "react";

import { AppSidebar } from "../components/app-sidebar";
import "./globals.css";

export const metadata: Metadata = {
  title: "Video Platform",
  description: "Example web consumer for the Video Platform API",
};

export default function RootLayout({ children }: Readonly<{ children: ReactNode }>) {
  return (
    <html lang="en">
      <body>
        <div className="min-h-screen lg:flex">
          <AppSidebar />
          <main className="min-w-0 flex-1 px-4 py-6 sm:px-6 sm:py-8 lg:px-10 lg:py-10">{children}</main>
        </div>
      </body>
    </html>
  );
}
