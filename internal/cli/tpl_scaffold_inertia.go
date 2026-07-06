// Inertia + Svelte frontend scaffold templates for cais new.
package cli

const tplAppHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0, viewport-fit=cover" />
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

createInertiaApp({
  resolve: (name) => {
    const pages = import.meta.glob('./pages/**/*.svelte', { eager: true })
    return pages[` + "`" + `./pages/${name}.svelte` + "`" + `]
  },
  setup({ el, App, props }) {
    new App({ target: el, props })
  },
})
`

const tplSvelteHome = `<script>
  import { inertia } from '@inertiajs/svelte'
  export let title = 'Home'
  export let site = {}
</script>

<div class="flex min-h-screen flex-col items-center justify-center px-6 py-14 text-center">
  <h1 class="mt-10 font-serif text-4xl font-semibold tracking-tight text-stone-800 md:text-5xl">
    You're on {{.AppName}} (Svelte + Inertia)!
  </h1>
  <p data-testid="inertia-ready">Svelte + Inertia ready</p>
  <p class="mt-3 max-w-md text-lg text-stone-600">Cais is ready to sail with Svelte.</p>
  <nav class="mt-6 space-x-4 text-sm">
    <!-- cais:nav -->
    <a href="/contact" use:inertia class="underline">Contact</a>
    <a href="/login" use:inertia class="underline">Login</a>
    <a href="/dashboard" use:inertia class="underline">Dashboard</a>
  </nav>
  <p class="mt-6 text-xs text-stone-500">Inertia component: {title}</p>
</div>
`

const tplSvelteContact = `<script>
  import { useForm } from '@inertiajs/svelte'
  export let errors = {}
  export let flash = {}
  export let site = {}
  let form = useForm({ name: '', email: '' })
  function submit() {
    $form.post('/contact')
  }
</script>

<div class="max-w-md mx-auto p-6">
  <h1 class="text-2xl font-semibold mb-4">Contact</h1>
  {#if flash.success}
    <p class="mb-4 text-green-700" data-testid="contact-success">{flash.success}</p>
  {/if}
  <form on:submit|preventDefault={submit}>
    <input type="text" bind:value={$form.name} placeholder="Name" class="block w-full border p-2" />
    {#if errors.name}<p class="text-red-600 text-sm">{errors.name}</p>{/if}
    <input type="email" bind:value={$form.email} placeholder="Email" class="block w-full border p-2 mt-2" />
    {#if errors.email}<p class="text-red-600 text-sm">{errors.email}</p>{/if}
    <button type="submit" class="mt-4 px-4 py-2 bg-amber-800 text-white">Send</button>
  </form>
</div>
`

const tplSvelteDashboard = `<script>
  import { inertia } from '@inertiajs/svelte'
  export let totalContacts = 0
  export let env = ''
  export let site = {}
</script>

<div class="p-8">
  <h1 class="text-3xl">Dashboard</h1>
  <p>Welcome! (Inertia/Svelte)</p>
  <p>Contacts: {totalContacts}</p>
  <p>Env: {env}</p>
  <form method="post" action="/logout" use:inertia>
    <button class="mt-4">Logout</button>
  </form>
</div>
`

const tplSvelteLogin = `<script>
  import { useForm } from '@inertiajs/svelte'
  export let errors = {}
  export let site = {}
  let form = useForm({ email: 'demo@example.com', password: 'password' })
  function submit() { $form.post('/login') }
</script>

<div class="max-w-sm mx-auto mt-10 p-6 border rounded">
  <h1 class="text-xl mb-4">Login</h1>
  <form on:submit|preventDefault={submit}>
    <input type="email" bind:value={$form.email} class="block w-full p-2 border" />
    {#if errors.email}<div class="text-red-600 text-xs">{errors.email}</div>{/if}
    <input type="password" bind:value={$form.password} class="block w-full p-2 border mt-2" />
    <button class="mt-4 bg-stone-800 text-white px-4 py-2">Log in</button>
  </form>
</div>
`

const tplSvelteSignup = `<script>
  import { useForm } from '@inertiajs/svelte'
  export let errors = {}
  let form = useForm({ email: '', password: '', password_confirmation: '' })
  function submit() { $form.post('/signup') }
</script>

<div class="max-w-sm mx-auto p-6">
  <h1>Sign up</h1>
  <form on:submit|preventDefault={submit}>
    <input bind:value={$form.email} type="email" placeholder="email" class="block w-full border p-2" />
    {#if errors.email}<p class="text-red-600">{errors.email}</p>{/if}
    <input bind:value={$form.password} type="password" class="block w-full border p-2 mt-2" />
    <input bind:value={$form.password_confirmation} type="password" class="block w-full border p-2 mt-2" />
    <button class="mt-4 px-4 py-2 bg-black text-white">Create account</button>
  </form>
</div>
`

const tplSvelteForgotPassword = `<script>
  import { useForm } from '@inertiajs/svelte'
  export let errors = {}
  let form = useForm({ email: '' })
  function submit() { $form.post('/forgot-password') }
</script>

<h1>Forgot password</h1>
<form on:submit|preventDefault={submit}>
  <input bind:value={$form.email} type="email" />
  {#if errors.email}<p>{errors.email}</p>{/if}
  <button>Send reset</button>
</form>
`

const tplSvelteResetPassword = `<script>
  import { useForm } from '@inertiajs/svelte'
  export let errors = {}
  export let token = ''
  let form = useForm({ token, password: '', password_confirmation: '' })
  function submit() { $form.post('/reset-password') }
</script>

<div class="max-w-sm mx-auto p-6">
  <h1 class="text-xl mb-4">Reset password</h1>
  {#if errors.token}<p class="text-red-600 text-sm mb-2">{errors.token}</p>{/if}
  <form on:submit|preventDefault={submit}>
    <input type="hidden" bind:value={$form.token} />
    <input bind:value={$form.password} type="password" class="block w-full border p-2" placeholder="New password" />
    {#if errors.password}<p class="text-red-600 text-sm">{errors.password}</p>{/if}
    <input bind:value={$form.password_confirmation} type="password" class="block w-full border p-2 mt-2" placeholder="Confirm password" />
    {#if errors.password_confirmation}<p class="text-red-600 text-sm">{errors.password_confirmation}</p>{/if}
    <button class="mt-4 px-4 py-2 bg-stone-800 text-white">Reset</button>
  </form>
</div>
`

const tplViteConfig = `import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import { resolve } from 'path'

export default defineConfig({
  plugins: [svelte()],
  root: '.',
  build: {
    manifest: true,
    outDir: 'web/static/build',
    emptyOutDir: false,
    rollupOptions: {
      input: resolve(__dirname, 'web/src/main.js'),
      output: {
        entryFileNames: 'assets/main.js',
        chunkFileNames: 'assets/[name].js',
        assetFileNames: 'assets/[name][extname]',
      },
    },
  },
  resolve: {
    alias: {
      '@': resolve(__dirname, 'web/src'),
    },
    conditions: ['browser'],
  },
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./vitest-setup.js'],
  },
})
`

const tplSvelteConfig = `import { vitePreprocess } from '@sveltejs/vite-plugin-svelte'

export default {
  preprocess: vitePreprocess(),
}
`

const tplVitestSetup = `import '@testing-library/jest-dom/vitest'
`

const tplBuildGitkeep = ``
