package cli

// Reference partial for HTMX SSE chat: history container + append-only SSE target.
const tplPartialChatSSE = `{{"{{"}}- define "chat_sse" -{{"}}"}}
<div id="chat-history" class="flex flex-col gap-3 min-h-[12rem]">
  {{"{{"}}- range .Messages {{"}}"}}
  <div class="rounded-xl px-4 py-2 max-w-[85%] {{"{{"}}if eq .Role "assistant"{{"}}"}}bg-white border border-slate-200 self-start{{"{{"}}else{{"}}"}}bg-indigo-600 text-white self-end{{"{{"}}end{{"}}"}}">
    <p class="text-sm whitespace-pre-wrap">{{"{{"}} .Content {{"}}"}}</p>
  </div>
  {{"{{"}}- end {{"}}"}}
</div>
<div
  id="chat-sse"
  hx-ext="sse"
  sse-connect="{{"{{"}} .StreamURL {{"}}"}}"
  sse-swap="message"
  hx-swap="beforeend"
  hx-target="#chat-history"
  class="hidden"
  aria-hidden="true"
></div>
{{"{{"}}- end -{{"}}"}}`
