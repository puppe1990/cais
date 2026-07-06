<script>
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
