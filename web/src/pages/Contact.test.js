import { render } from "@testing-library/svelte";
import { writable } from "svelte/store";
import { describe, it, expect, vi } from "vitest";

vi.mock("@inertiajs/svelte", () => ({
  useForm: (data) =>
    writable({
      ...data,
      errors: {},
      processing: false,
      recentlySuccessful: false,
      post: vi.fn(),
    }),
  inertia: () => {},
}));

import Contact from "./Contact.svelte";

describe("Contact Svelte component (Inertia page)", () => {
  it("renders the contact heading", () => {
    const { getByText } = render(Contact, {
      props: { errors: {}, flash: {}, site: {} },
    });
    expect(getByText("Contact")).toBeTruthy();
  });

  it("shows success message from flash prop", () => {
    const { getByTestId } = render(Contact, {
      props: {
        errors: {},
        flash: { success: "Message sent successfully." },
        site: {},
      },
    });
    expect(getByTestId("contact-success").textContent).toContain("Message sent successfully.");
  });
});
