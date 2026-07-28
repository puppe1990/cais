<script>
  import { useForm } from '@inertiajs/svelte'
  import PasswordInput from '../components/PasswordInput.svelte'
  export let errors = {}
  export let site = {}
  let form = useForm({ email: 'demo@example.com', password: 'password' })
  function submit() { form.post('/login') }
</script>

<svelte:head>
  <title>Login · Cais</title>
</svelte:head>

<div class="max-w-sm mx-auto mt-10 p-6 border rounded">
  <h1 class="text-xl mb-4">Login</h1>
  <form on:submit|preventDefault={submit}>
    <input type="email" bind:value={form.email} class="block w-full p-2 border" />
    {#if errors.email}<div class="text-red-600 text-xs">{errors.email}</div>{/if}
    <div class="mt-2">
      <PasswordInput bind:value={form.password} name="password" autocomplete="current-password" className="block w-full p-2 border" />
    </div>
    <button class="mt-4 bg-stone-800 text-white px-4 py-2">Log in</button>
  </form>
</div>
