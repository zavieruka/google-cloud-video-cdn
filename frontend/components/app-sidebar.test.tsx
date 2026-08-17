import { render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

const usePathname = vi.hoisted(() => vi.fn());

vi.mock("next/navigation", () => ({ usePathname }));

import { AppSidebar } from "./app-sidebar";

describe("AppSidebar", () => {
  afterEach(() => {
    usePathname.mockReset();
  });

  it("keeps Videos active while viewing a video detail page", () => {
    usePathname.mockReturnValue("/videos/video-123");

    render(<AppSidebar />);

    expect(screen.getByRole("navigation", { name: "Main navigation" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Videos" })).toHaveAttribute("aria-current", "page");
    expect(screen.getByRole("link", { name: "Upload" })).not.toHaveAttribute("aria-current");
  });
});
