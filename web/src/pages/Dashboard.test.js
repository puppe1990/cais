import { render, fireEvent } from "@testing-library/svelte";
import { describe, it, expect, vi } from "vitest";

const { post } = vi.hoisted(() => ({ post: vi.fn() }));
vi.mock("@inertiajs/svelte", () => ({
  router: { post },
  inertia: () => {},
}));

import Dashboard from "./Dashboard.svelte";

describe("Dashboard Svelte component (Inertia page)", () => {
  it("renders contact count from props", () => {
    const { getByText } = render(Dashboard, {
      props: { totalContacts: 42, env: "test", site: {} },
    });
    expect(getByText("Contacts: 42")).toBeTruthy();
  });

  it("posts logout via router.post", async () => {
    post.mockClear();
    const { getByText } = render(Dashboard, {
      props: { totalContacts: 0, env: "test", site: {} },
    });
    await fireEvent.click(getByText("Logout"));
    expect(post).toHaveBeenCalledWith("/logout");
  });
});
