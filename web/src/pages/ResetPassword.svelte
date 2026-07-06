<script>
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
