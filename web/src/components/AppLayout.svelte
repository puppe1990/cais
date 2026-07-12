<script>
  import { inertia } from '@inertiajs/svelte'
  export let site = {}
  export let flash = {}
  export let labels = {}
</script>

<div class="min-h-screen flex flex-col bg-stone-50 text-stone-900">
  <header class="border-b border-stone-200 bg-white/80 backdrop-blur">
    <nav class="mx-auto flex max-w-4xl items-center justify-between px-4 py-3 text-sm">
      <a href="/" use:inertia class="font-semibold text-stone-800">{site.appName || 'Cais'}</a>
      <div class="flex gap-4">
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