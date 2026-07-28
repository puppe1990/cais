<script>
  import { useForm } from '@inertiajs/svelte'
  import PasswordInput from '../components/PasswordInput.svelte'
  export let errors = {}
  export let token = ''
  let form = useForm({ token, password: '', password_confirmation: '' })
  function submit() { form.post('/reset-password') }
</script>

<svelte:head>
  <title>Reset password · Cais</title>
</svelte:head>

<div class="max-w-sm mx-auto p-6">
  <h1 class="text-xl mb-4">Reset password</h1>
  {#if errors.token}<p class="text-red-600 text-sm mb-2">{errors.token}</p>{/if}
  <form on:submit|preventDefault={submit}>
    <input type="hidden" bind:value={form.token} />
    <PasswordInput bind:value={form.password} name="password" autocomplete="new-password" placeholder="New password" className="block w-full border p-2" />
    {#if errors.password}<p class="text-red-600 text-sm">{errors.password}</p>{/if}
    <div class="mt-2">
      <PasswordInput bind:value={form.password_confirmation} name="password_confirmation" autocomplete="new-password" placeholder="Confirm password" className="block w-full border p-2" />
    </div>
    {#if errors.password_confirmation}<p class="text-red-600 text-sm">{errors.password_confirmation}</p>{/if}
    <button class="mt-4 px-4 py-2 bg-stone-800 text-white">Reset</button>
  </form>
</div>
