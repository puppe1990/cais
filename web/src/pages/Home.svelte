<script>
  import AppLayout from '../components/AppLayout.svelte'
  export let title = 'Home'
  export let site = {}
  export let flash = {}
  export let labels = {}
</script>

<AppLayout {site} {flash} {labels}>
  <div class="flex flex-col items-center justify-center px-6 py-14 text-center">
    <h1 class="mt-10 font-serif text-4xl font-semibold tracking-tight text-stone-800 md:text-5xl">
      {labels.heading || "You're on Cais!"}
    </h1>
    <p data-testid="inertia-ready" class="mt-3 text-lg text-stone-600">{labels.subtitle || 'Cais is ready to sail with Svelte.'}</p>
    <p class="mt-2 text-sm text-stone-500">{labels.stack || 'Go · Inertia · Svelte · SQLite'}</p>
    <p class="mt-6 text-xs text-stone-500">Inertia component: {title}</p>
  </div>
</AppLayout>