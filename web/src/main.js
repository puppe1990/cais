import { createInertiaApp } from "@inertiajs/svelte";
import { mount } from "svelte";

createInertiaApp({
  resolve: (name) => {
    const pages = import.meta.glob("./pages/**/*.svelte", { eager: true });
    const page = pages[`./pages/${name}.svelte`];
    if (!page) throw new Error(`Inertia page not found: ${name}`);
    return page;
  },
  setup({ el, App, props }) {
    mount(App, { target: el, props });
  },
  // Match pkg/cais/csrf double-submit cookie (not Laravel XSRF defaults).
  http: {
    xsrfCookieName: "cais_csrf",
    xsrfHeaderName: "X-CSRF-Token",
  },
});
