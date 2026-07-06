import { render } from "@testing-library/svelte";
import { describe, it, expect } from "vitest";
import Dashboard from "./Dashboard.svelte";

describe("Dashboard Svelte component (Inertia page)", () => {
  it("renders contact count from props", () => {
    const { getByText } = render(Dashboard, {
      props: { totalContacts: 42, env: "test", site: {} },
    });
    expect(getByText("Contacts: 42")).toBeTruthy();
  });
});
