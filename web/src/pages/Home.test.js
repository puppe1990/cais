import { render } from "@testing-library/svelte";
import { describe, it, expect } from "vitest";
import Home from "./Home.svelte";

describe("Home Svelte component (Inertia page)", () => {
  it("renders the welcome heading from labels prop", () => {
    const { getByText } = render(Home, {
      props: {
        title: "Home",
        site: { appName: "Cais" },
        labels: { heading: "You're on Cais!" },
      },
    });
    expect(getByText(/You're on Cais/i)).toBeTruthy();
  });
});
