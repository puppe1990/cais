import { render, fireEvent } from "@testing-library/svelte";
import { describe, it, expect } from "vitest";
import PasswordInput from "./PasswordInput.svelte";

describe("PasswordInput", () => {
  it("renders a password input with toggle control", () => {
    const { container, getByLabelText } = render(PasswordInput, {
      props: { name: "password", value: "secret" },
    });
    const input = container.querySelector('input[name="password"]');
    expect(input).toBeTruthy();
    expect(input.type).toBe("password");
    expect(container.querySelector(".cais-password-wrap")).toBeTruthy();
    expect(getByLabelText("Show password")).toBeTruthy();
  });

  it("toggles visibility on eye click", async () => {
    const { container, getByLabelText } = render(PasswordInput, {
      props: { name: "password", value: "secret" },
    });
    const input = container.querySelector('input[name="password"]');
    const btn = getByLabelText("Show password");
    await fireEvent.click(btn);
    expect(input.type).toBe("text");
    expect(getByLabelText("Hide password")).toBeTruthy();
    await fireEvent.click(getByLabelText("Hide password"));
    expect(input.type).toBe("password");
  });
});
