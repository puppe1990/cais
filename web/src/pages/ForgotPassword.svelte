<script>
  import { useForm } from '@inertiajs/svelte'
  export let errors = {}
  let form = useForm({ email: '' })
  function submit() { $form.post('/forgot-password') }
</script>

<h1>Forgot password</h1>
<form on:submit|preventDefault={submit}>
  <input bind:value={$form.email} type="email" />
  {#if errors.email}<p>{errors.email}</p>{/if}
  <button>Send reset</button>
</form>
