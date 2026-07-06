<script>
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
