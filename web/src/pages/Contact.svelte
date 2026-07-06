<script>
  import { useForm } from '@inertiajs/svelte'
  export let errors = {}
  export let flash = {}
  export let site = {}
  let form = useForm({ name: '', email: '' })
  function submit() {
    $form.post('/contact')
  }
</script>

<div class="max-w-md mx-auto p-6">
  <h1 class="text-2xl font-semibold mb-4">Contact</h1>
  {#if flash.success}
    <p class="mb-4 text-green-700" data-testid="contact-success">{flash.success}</p>
  {/if}
  <form on:submit|preventDefault={submit}>
    <input type="text" bind:value={$form.name} placeholder="Name" class="block w-full border p-2" />
    {#if errors.name}<p class="text-red-600 text-sm">{errors.name}</p>{/if}
    <input type="email" bind:value={$form.email} placeholder="Email" class="block w-full border p-2 mt-2" />
    {#if errors.email}<p class="text-red-600 text-sm">{errors.email}</p>{/if}
    <button type="submit" class="mt-4 px-4 py-2 bg-amber-800 text-white">Send</button>
  </form>
</div>
