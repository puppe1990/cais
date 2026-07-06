<script>
  import { inertia } from '@inertiajs/svelte'
  export let title = 'Home'
  export let site = {}
</script>

<div class="flex min-h-screen flex-col items-center justify-center px-6 py-14 text-center">
  <h1 class="mt-10 font-serif text-4xl font-semibold tracking-tight text-stone-800 md:text-5xl">
    You're on Cais (Svelte + Inertia)!
  </h1>
  <p data-testid="inertia-ready">Svelte + Inertia TDD ready</p>
  <p class="mt-3 max-w-md text-lg text-stone-600">Cais is ready to sail with Svelte.</p>
  <nav class="mt-6 space-x-4 text-sm">
    <a href="/contact" use:inertia class="underline">Contact</a>
    <a href="/login" use:inertia class="underline">Login</a>
    <a href="/dashboard" use:inertia class="underline">Dashboard</a>
  </nav>
  <p class="mt-6 text-xs text-stone-500">Inertia component: {title}</p>
</div>
