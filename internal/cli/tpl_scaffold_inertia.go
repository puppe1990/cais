// Inertia + Svelte frontend scaffold templates for cais new.
package cli

const tplAppHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0, viewport-fit=cover" />
  <title>{{.AppName}}</title>
  {{"{{ .inertiaHead }}"}}
  <link rel="stylesheet" href="/static/css/styles.css" />
  <link rel="manifest" href="/static/manifest.webmanifest" />
  <meta name="theme-color" content="#4f46e5" />
  <link rel="icon" href="/static/icons/icon.png" type="image/png" />
</head>
<body>
  {{"{{ .inertia }}"}}
  <script type="module" src="/static/build/assets/main.js"></script>
</body>
</html>
`

const tplMainJS = `import { createInertiaApp } from '@inertiajs/svelte'
import { mount } from 'svelte'

createInertiaApp({
  resolve: (name) => {
    const pages = import.meta.glob('./pages/**/*.svelte', { eager: true })
    const page = pages[` + "`" + `./pages/${name}.svelte` + "`" + `]
    if (!page) throw new Error(` + "`" + `Inertia page not found: ${name}` + "`" + `)
    return page
  },
  setup({ el, App, props }) {
    mount(App, { target: el, props })
  },
  // Match pkg/cais/csrf double-submit cookie (not Laravel XSRF defaults).
  // Production sets Secure cookies, which use the __Host- prefix (#173);
  // Vite's PROD flag tracks ENV=production, so pick the matching name.
  http: {
    xsrfCookieName: import.meta.env.PROD ? '__Host-cais_csrf' : 'cais_csrf',
    xsrfHeaderName: 'X-CSRF-Token',
  },
})
`

const tplAppLayout = `<script>
  import { inertia } from '@inertiajs/svelte'
  export let site = {}
  export let flash = {}
  export let labels = {}
</script>

<div class="min-h-screen flex flex-col bg-stone-50 text-stone-900">
  <header class="border-b border-stone-200 bg-white/80 backdrop-blur">
    <nav class="mx-auto flex max-w-4xl items-center justify-between px-4 py-3 text-sm">
      <a href="/" use:inertia class="font-semibold text-stone-800">{site.appName || '{{.AppName}}'}</a>
      <div class="flex gap-4">
        <!-- cais:nav -->
        <a href="/contact" use:inertia class="hover:text-stone-600">{labels.contact || 'Contact'}</a>
        <a href="/login" use:inertia class="hover:text-stone-600">{labels.login || 'Login'}</a>
        <a href="/dashboard" use:inertia class="hover:text-stone-600">{labels.dashboard || 'Dashboard'}</a>
      </div>
    </nav>
  </header>

  {#if flash.notice}
    <p class="mx-auto mt-4 max-w-4xl rounded-lg bg-green-50 px-4 py-2 text-sm text-green-800" data-testid="flash-notice">{flash.notice}</p>
  {/if}
  {#if flash.success}
    <p class="mx-auto mt-4 max-w-4xl rounded-lg bg-green-50 px-4 py-2 text-sm text-green-800" data-testid="flash-success">{flash.success}</p>
  {/if}

  <main class="flex-1">
    <slot />
  </main>
</div>
`

const tplSvelteHome = `<script>
  import AppLayout from '../components/AppLayout.svelte'
  export let title = 'Home'
  export let site = {}
  export let flash = {}
  export let labels = {}
</script>

<svelte:head>
  <title>{title} · {{.AppName}}</title>
</svelte:head>

<AppLayout {site} {flash} {labels}>
  <div class="flex flex-col items-center justify-center px-6 py-14 text-center">
    <h1 class="mt-10 font-serif text-4xl font-semibold tracking-tight text-stone-800 md:text-5xl">
      {labels.heading || "You're on {{.AppName}}!"}
    </h1>
    <p data-testid="inertia-ready" class="mt-3 text-lg text-stone-600">{labels.subtitle || '{{.AppName}} is ready to sail.'}</p>
    <p class="mt-2 text-sm text-stone-500">{labels.stack || 'Go · Inertia · Svelte · SQLite'}</p>
    <p class="mt-6 text-xs text-stone-500">Inertia component: {title}</p>
  </div>
</AppLayout>
`

const tplSvelteContact = `<script>
  import { useForm } from '@inertiajs/svelte'
  export let errors = {}
  export let flash = {}
  export let site = {}
  // useForm is not a store (no $form). Avoid $: form.x = prop — can blank the page under Svelte 5.
  let form = useForm({ name: '', email: '' })
  function submit() {
    form.post('/contact')
  }
</script>

<svelte:head>
  <title>Contact · {{.AppName}}</title>
</svelte:head>

<div class="max-w-md mx-auto p-6">
  <h1 class="text-2xl font-semibold mb-4">Contact</h1>
  {#if flash.success}
    <p class="mb-4 text-green-700" data-testid="contact-success">{flash.success}</p>
  {/if}
  <form on:submit|preventDefault={submit}>
    <input type="text" bind:value={form.name} placeholder="Name" class="block w-full border p-2" />
    {#if errors.name}<p class="text-red-600 text-sm">{errors.name}</p>{/if}
    <input type="email" bind:value={form.email} placeholder="Email" class="block w-full border p-2 mt-2" />
    {#if errors.email}<p class="text-red-600 text-sm">{errors.email}</p>{/if}
    <button type="submit" class="mt-4 px-4 py-2 bg-amber-800 text-white">Send</button>
  </form>
</div>
`

const tplSvelteDashboard = `<script>
  import { router } from '@inertiajs/svelte'
  export let totalContacts = 0
  export let env = ''
  export let site = {}
  export let flash = {}
  // use:inertia is for GET navigation; mutations need router.post / useForm.
  function logout() { router.post('/logout') }
</script>

<svelte:head>
  <title>Dashboard · {{.AppName}}</title>
</svelte:head>

<div class="p-8">
  <h1 class="text-3xl">Dashboard</h1>
  {#if flash.notice}
    <p class="mb-4 rounded-lg bg-green-50 px-4 py-2 text-sm text-green-800" data-testid="flash-notice">{flash.notice}</p>
  {/if}
  <p>Welcome! (Inertia/Svelte)</p>
  <p>Contacts: {totalContacts}</p>
  <p>Env: {env}</p>
  <button type="button" class="mt-4" on:click={logout}>Logout</button>
</div>
`

// tplAuthLayout centers auth forms on the viewport (login/signup/reset default shell).
const tplAuthLayout = `<script>
  export let site = {}
  export let flash = {}
  export let title = ''
</script>

<div class="flex min-h-screen items-center justify-center bg-stone-50 px-4 py-10 text-stone-900">
  <div class="w-full max-w-sm">
    {#if site.appName}
      <p class="mb-4 text-center text-sm font-semibold tracking-wide text-stone-600">{site.appName}</p>
    {/if}
    <div class="rounded-2xl border border-stone-200 bg-white p-6 shadow-sm">
      {#if title}
        <h1 class="mb-4 text-center text-xl font-semibold text-stone-800">{title}</h1>
      {/if}
      {#if flash.notice}
        <p class="mb-4 rounded-lg bg-green-50 px-4 py-2 text-sm text-green-800" data-testid="flash-notice">{flash.notice}</p>
      {/if}
      {#if flash.success}
        <p class="mb-4 rounded-lg bg-green-50 px-4 py-2 text-sm text-green-800" data-testid="flash-success">{flash.success}</p>
      {/if}
      <slot />
    </div>
  </div>
</div>
`

const tplPasswordInput = `<script>
  /** Shared password field with eye show/hide — use for every password input. */
  export let value = ''
  export let name = 'password'
  export let id = ''
  export let placeholder = ''
  export let required = false
  export let autocomplete = 'current-password'
  export let className = 'block w-full border p-2'
  let visible = false
  $: inputId = id || name
  function toggle() { visible = !visible }
</script>

<div class="cais-password-wrap">
  <input type={visible ? 'text' : 'password'} {name} id={inputId} {placeholder} {required} {autocomplete} class={className} bind:value />
  <button type="button" class="cais-password-toggle" data-cais-password-toggle aria-label={visible ? 'Hide password' : 'Show password'} aria-pressed={visible} on:click={toggle}>
    {#if visible}
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" /></svg>
    {:else}
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" /></svg>
    {/if}
  </button>
</div>
`

const tplSvelteLogin = `<script>
  import { useForm } from '@inertiajs/svelte'
  import AuthLayout from '../components/AuthLayout.svelte'
  import PasswordInput from '../components/PasswordInput.svelte'
  export let errors = {}
  export let site = {}
  export let flash = {}
  let form = useForm({ email: 'demo@example.com', password: 'password' })
  function submit() { form.post('/login') }
</script>

<svelte:head>
  <title>Login · {{.AppName}}</title>
</svelte:head>

<AuthLayout {site} {flash} title="Log in">
  <form on:submit|preventDefault={submit} class="space-y-3">
    <div>
      <input type="email" bind:value={form.email} placeholder="Email" class="block w-full rounded-lg border border-stone-300 p-2.5 text-sm" />
      {#if errors.email}<p class="mt-1 text-xs text-red-600">{errors.email}</p>{/if}
    </div>
    <div>
      <PasswordInput bind:value={form.password} name="password" autocomplete="current-password" className="block w-full rounded-lg border border-stone-300 p-2.5 text-sm" />
    </div>
    <button type="submit" class="w-full rounded-lg bg-stone-800 px-4 py-2.5 text-sm font-medium text-white hover:bg-stone-900">Log in</button>
    <p class="pt-1 text-center text-xs text-stone-500">
      <a href="/signup" class="text-stone-700 underline hover:text-stone-900">Create account</a>
      ·
      <a href="/forgot-password" class="text-stone-700 underline hover:text-stone-900">Forgot password</a>
    </p>
  </form>
</AuthLayout>
`

const tplSvelteSignup = `<script>
  import { useForm } from '@inertiajs/svelte'
  import AuthLayout from '../components/AuthLayout.svelte'
  import PasswordInput from '../components/PasswordInput.svelte'
  export let errors = {}
  export let site = {}
  export let flash = {}
  let form = useForm({ email: '', password: '', password_confirmation: '' })
  function submit() { form.post('/signup') }
</script>

<svelte:head>
  <title>Sign up · {{.AppName}}</title>
</svelte:head>

<AuthLayout {site} {flash} title="Create account">
  <form on:submit|preventDefault={submit} class="space-y-3">
    <div>
      <input bind:value={form.email} type="email" placeholder="Email" class="block w-full rounded-lg border border-stone-300 p-2.5 text-sm" />
      {#if errors.email}<p class="mt-1 text-xs text-red-600">{errors.email}</p>{/if}
    </div>
    <div>
      <PasswordInput bind:value={form.password} name="password" autocomplete="new-password" className="block w-full rounded-lg border border-stone-300 p-2.5 text-sm" />
      {#if errors.password}<p class="mt-1 text-xs text-red-600">{errors.password}</p>{/if}
    </div>
    <div>
      <PasswordInput bind:value={form.password_confirmation} name="password_confirmation" autocomplete="new-password" className="block w-full rounded-lg border border-stone-300 p-2.5 text-sm" />
      {#if errors.password_confirmation}<p class="mt-1 text-xs text-red-600">{errors.password_confirmation}</p>{/if}
    </div>
    <button type="submit" class="w-full rounded-lg bg-stone-800 px-4 py-2.5 text-sm font-medium text-white hover:bg-stone-900">Create account</button>
    <p class="pt-1 text-center text-xs text-stone-500">
      <a href="/login" class="text-stone-700 underline hover:text-stone-900">Already have an account?</a>
    </p>
  </form>
</AuthLayout>
`

const tplSvelteForgotPassword = `<script>
  import { useForm } from '@inertiajs/svelte'
  import AuthLayout from '../components/AuthLayout.svelte'
  export let errors = {}
  export let site = {}
  export let flash = {}
  let form = useForm({ email: '' })
  function submit() { form.post('/forgot-password') }
</script>

<svelte:head>
  <title>Forgot password · {{.AppName}}</title>
</svelte:head>

<AuthLayout {site} {flash} title="Forgot password">
  <form on:submit|preventDefault={submit} class="space-y-3">
    <div>
      <input bind:value={form.email} type="email" placeholder="Email" class="block w-full rounded-lg border border-stone-300 p-2.5 text-sm" />
      {#if errors.email}<p class="mt-1 text-xs text-red-600">{errors.email}</p>{/if}
    </div>
    <button type="submit" class="w-full rounded-lg bg-stone-800 px-4 py-2.5 text-sm font-medium text-white hover:bg-stone-900">Send reset link</button>
    <p class="pt-1 text-center text-xs text-stone-500">
      <a href="/login" class="text-stone-700 underline hover:text-stone-900">Back to login</a>
    </p>
  </form>
</AuthLayout>
`

const tplSvelteResetPassword = `<script>
  import { useForm } from '@inertiajs/svelte'
  import AuthLayout from '../components/AuthLayout.svelte'
  import PasswordInput from '../components/PasswordInput.svelte'
  export let errors = {}
  export let site = {}
  export let flash = {}
  export let token = ''
  let form = useForm({ token, password: '', password_confirmation: '' })
  function submit() { form.post('/reset-password') }
</script>

<svelte:head>
  <title>Reset password · {{.AppName}}</title>
</svelte:head>

<AuthLayout {site} {flash} title="Reset password">
  {#if errors.token}
    <p class="mb-3 text-sm text-red-600">{errors.token}</p>
  {/if}
  <form on:submit|preventDefault={submit} class="space-y-3">
    <input type="hidden" bind:value={form.token} />
    <div>
      <PasswordInput bind:value={form.password} name="password" autocomplete="new-password" placeholder="New password" className="block w-full rounded-lg border border-stone-300 p-2.5 text-sm" />
      {#if errors.password}<p class="mt-1 text-xs text-red-600">{errors.password}</p>{/if}
    </div>
    <div>
      <PasswordInput bind:value={form.password_confirmation} name="password_confirmation" autocomplete="new-password" placeholder="Confirm password" className="block w-full rounded-lg border border-stone-300 p-2.5 text-sm" />
      {#if errors.password_confirmation}<p class="mt-1 text-xs text-red-600">{errors.password_confirmation}</p>{/if}
    </div>
    <button type="submit" class="w-full rounded-lg bg-stone-800 px-4 py-2.5 text-sm font-medium text-white hover:bg-stone-900">Reset password</button>
  </form>
</AuthLayout>
`

const tplViteConfig = `import { defineConfig } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import { resolve } from "path";

export default defineConfig({
  plugins: [svelte()],
  root: ".",
  build: {
    manifest: true,
    outDir: "web/static/build",
    emptyOutDir: false,
    rollupOptions: {
      input: resolve(__dirname, "web/src/main.js"),
      output: {
        entryFileNames: "assets/main.js",
        chunkFileNames: "assets/[name].js",
        assetFileNames: "assets/[name][extname]",
      },
    },
  },
  resolve: {
    alias: {
      "@": resolve(__dirname, "web/src"),
    },
    conditions: ["browser"],
  },
  test: {
    environment: "jsdom",
    globals: true,
    setupFiles: ["./vitest-setup.js"],
  },
});
`

const tplSvelteConfig = `import { vitePreprocess } from "@sveltejs/vite-plugin-svelte";

export default {
  preprocess: vitePreprocess(),
};
`

const tplVitestSetup = `import "@testing-library/jest-dom/vitest";
`

const tplBuildGitkeep = ``
