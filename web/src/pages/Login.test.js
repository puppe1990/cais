import { render } from "@testing-library/svelte";
import { writable } from "svelte/store";
import { describe, it, expect, vi } from "vitest";

vi.mock("@inertiajs/svelte", () => ({
  useForm: (data) =>
    writable({
      ...data,
      errors: {},
      processing: false,
      post: vi.fn(),
    }),
  inertia: () => {},
}));

import Login from "./Login.svelte";

describe("Login Svelte component (Inertia page)", () => {
  it("renders the login heading", () => {
    const { getByText } = render(Login, {
      props: { errors: {}, site: {} },
    });
    expect(getByText("Login")).toBeTruthy();
  });
});
