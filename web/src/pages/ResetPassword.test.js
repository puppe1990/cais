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

import ResetPassword from "./ResetPassword.svelte";

describe("ResetPassword Svelte component (Inertia page)", () => {
  it("renders token validation error from props.errors", () => {
    const { getByText } = render(ResetPassword, {
      props: {
        errors: { token: "Invalid or expired reset token." },
        token: "",
      },
    });
    expect(getByText("Invalid or expired reset token.")).toBeTruthy();
  });
});
