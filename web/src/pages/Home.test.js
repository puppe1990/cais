import { render } from "@testing-library/svelte";
import { describe, it, expect } from "vitest";
import Home from "./Home.svelte";

describe("Home Svelte component (Inertia page)", () => {
  it("renders the welcome heading from props", () => {
    const { getByText } = render(Home, {
      props: { title: "Home", site: {} },
    });
    expect(getByText(/You're on Cais/i)).toBeTruthy();
  });
});
