<script>
  import { useForm } from '@inertiajs/svelte'
  import PasswordInput from '../components/PasswordInput.svelte'
  export let errors = {}
  let form = useForm({ email: '', password: '', password_confirmation: '' })
  function submit() { form.post('/signup') }
</script>

<svelte:head>
  <title>Sign up · Cais</title>
</svelte:head>

<div class="max-w-sm mx-auto p-6">
  <h1>Sign up</h1>
  <form on:submit|preventDefault={submit}>
    <input bind:value={form.email} type="email" placeholder="email" class="block w-full border p-2" />
    {#if errors.email}<p class="text-red-600">{errors.email}</p>{/if}
    <div class="mt-2">
      <PasswordInput bind:value={form.password} name="password" autocomplete="new-password" className="block w-full border p-2" />
    </div>
    <div class="mt-2">
      <PasswordInput bind:value={form.password_confirmation} name="password_confirmation" autocomplete="new-password" className="block w-full border p-2" />
    </div>
    <button class="mt-4 px-4 py-2 bg-black text-white">Create account</button>
  </form>
</div>
