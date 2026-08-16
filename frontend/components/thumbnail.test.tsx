import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { Thumbnail } from "./thumbnail";

describe("Thumbnail", () => {
  it("crops the selected candidate from the four-column sheet", () => {
    render(<Thumbnail alt="Demo thumbnail" index={5} url="https://example.test/sheet.jpeg" />);

    const thumbnail = screen.getByRole("img", { name: "Demo thumbnail" });
    expect(thumbnail.style.backgroundImage).toBe('url("https://example.test/sheet.jpeg")');
    const [horizontalPosition, verticalPosition] = thumbnail.style.backgroundPosition.split(" ");
    expect(Number.parseFloat(horizontalPosition)).toBeCloseTo(100 / 3);
    expect(verticalPosition).toBe("50%");
    expect(thumbnail.style.backgroundSize).toBe("400% 300%");
  });
});
