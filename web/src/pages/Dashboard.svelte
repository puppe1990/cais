<script>
  import { router } from '@inertiajs/svelte'
  export let totalContacts = 0
  export let env = ''
  export let site = {}
  // use:inertia is for GET navigation; mutations need router.post / useForm.
  function logout() { router.post('/logout') }
</script>

<svelte:head>
  <title>Dashboard · Cais</title>
</svelte:head>

<div class="p-8">
  <h1 class="text-3xl">Dashboard</h1>
  <p>Welcome! (Inertia/Svelte)</p>
  <p>Contacts: {totalContacts}</p>
  <p>Env: {env}</p>
  <button type="button" class="mt-4" on:click={logout}>Logout</button>
</div>
