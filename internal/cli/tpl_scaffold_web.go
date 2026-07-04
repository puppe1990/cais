package cli

// Shared base layout fragments for cais new (full, minimal, blank). Edit fragments once;
// tplLayout / tplLayoutMinimal / tplLayoutBlank compose the generated base.html variants.
const tplLayoutTitleDesc = `{{"{{"}} define "title" {{"}}"}}{{.AppName}}{{"{{"}} end {{"}}"}}
{{"{{"}} define "description" {{"}}"}}{{.AppName}} — powered by Cais{{"{{"}} end {{"}}"}}`

const tplLayoutBaseOpen = `{{"{{"}} define "base" {{"}}"}}
<!doctype html>
<html lang="{{"{{"}} htmlLang {{"}}"}}">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0, viewport-fit=cover" />
    {{"{{"}} if .CSRFToken {{"}}"}}<meta name="csrf-token" content="{{"{{"}} .CSRFToken {{"}}"}}" />{{"{{"}} end {{"}}"}}
    <title>{{"{{"}} template "title" . {{"}}"}}</title>
    <meta name="description" content="{{"{{"}} template "description" . {{"}}"}}" />
    <meta property="og:type" content="website" />
    <meta property="og:site_name" content="{{.AppName}}" />
    <meta property="og:title" content="{{"{{"}} template "title" . {{"}}"}}" />
    <meta property="og:description" content="{{"{{"}} template "description" . {{"}}"}}" />
    <meta property="og:image" content="{{"{{"}} absURL .AppURL "/static/og.png" {{"}}"}}" />
    <meta property="og:locale" content="{{"{{"}} ogLocale {{"}}"}}" />
    <meta name="twitter:card" content="summary_large_image" />
    <meta name="twitter:title" content="{{"{{"}} template "title" . {{"}}"}}" />
    <meta name="twitter:description" content="{{"{{"}} template "description" . {{"}}"}}" />
    <meta name="twitter:image" content="{{"{{"}} absURL .AppURL "/static/og.png" {{"}}"}}" />
    <link rel="stylesheet" href="/static/css/styles.css" />
    <link rel="manifest" href="/static/manifest.webmanifest" />
    <meta name="theme-color" content="#4f46e5" />
    <meta name="mobile-web-app-capable" content="yes" />
    <meta name="apple-mobile-web-app-capable" content="yes" />
    <meta name="apple-mobile-web-app-status-bar-style" content="black-translucent" />
    <meta name="apple-mobile-web-app-title" content="{{.AppName}}" />
    <link rel="apple-touch-icon" href="/static/icons/icon.png" />
    <link rel="icon" href="/static/icons/icon.png" type="image/png" />
    <script src="/static/js/htmx.min.js" defer></script>
    <script src="/static/js/idiomorph-ext.min.js" defer></script>
    <script src="/static/js/sse-ext.min.js" defer></script>
    <script src="/static/js/cais.js" defer></script>
  </head>
  <body hx-ext="morph,sse" class="min-h-screen bg-slate-50 font-sans antialiased text-slate-900 flex flex-col justify-between">
    <div>
      <header class="bg-white border-b border-slate-200 sticky top-0 z-40 shadow-xs">
        <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-2.5 flex flex-col md:flex-row md:items-center md:justify-between gap-3">
          <a href="/" class="flex items-center gap-2.5 group">
            <div class="p-2 bg-indigo-600 rounded-lg text-white shadow-xs flex items-center justify-center group-hover:bg-indigo-700 transition">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 11-4 0 2 2 0 014 0z" />
              </svg>
            </div>
            <div>
              <h1 class="text-lg font-black text-slate-900 tracking-tight font-display flex items-center gap-1.5 leading-none">
                {{.AppName}}
                <span class="text-[9px] bg-indigo-100 text-indigo-800 px-1.5 py-0.5 rounded-md font-bold uppercase tracking-wider">Beta</span>
              </h1>
              <p class="text-[10px] text-slate-500 font-semibold mt-1">Powered by Cais</p>
            </div>
          </a>
        </div>
      </header>
      <nav id="cais-nav" class="bg-white border-b border-slate-200 shadow-2xs sticky top-[53px] z-30">
        <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div class="flex space-x-1 py-1.5 overflow-x-auto no-scrollbar">
            `

const tplLayoutNavFull = `<!-- cais:nav -->
            {{"{{"}} template "nav_links" . {{"}}"}}`

const tplLayoutNavEmpty = `<!-- cais:nav -->`

const tplLayoutBaseClose = `
          </div>
        </div>
      </nav>
      <div id="cais-toast-host" aria-live="polite">
        {{"{{"}} if .Flash {{"}}"}}
        <div class="fixed top-24 left-1/2 -translate-x-1/2 z-50 bg-slate-900 text-white px-5 py-3 rounded-2xl shadow-xl flex items-center gap-2 border border-slate-700/50" role="status">
          {{"{{"}} template "icon_sparkles_md" . {{"}}"}}
          <span class="text-xs font-bold">{{"{{"}} .Flash {{"}}"}}</span>
        </div>
        {{"{{"}} end {{"}}"}}
      </div>
      <main id="cais-main" data-cais-view-transition class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-5 flex-grow">{{"{{"}} template "content" . {{"}}"}}</main>
    </div>
    <footer class="mt-auto border-t border-slate-200/80 pt-8 pb-6 text-center text-xs text-slate-400">
      <div class="max-w-7xl mx-auto px-4">
        <p>© 2026 {{.AppName}}. Built with Cais.</p>
        <p class="mt-1">HTMX + Go + SQLite — server-rendered, app-like UX.</p>
      </div>
    </footer>
    {{"{{"}} if eq .Env "development" {{"}}"}}
    <script>
      if ("serviceWorker" in navigator) {
        navigator.serviceWorker.getRegistrations().then(function (regs) {
          regs.forEach(function (r) { r.unregister(); });
        });
        if ("caches" in window) {
          caches.keys().then(function (keys) {
            keys.forEach(function (k) { caches.delete(k); });
          });
        }
      }
    </script>
    {{"{{"}} else {{"}}"}}
    <script>
      if ("serviceWorker" in navigator) {
        navigator.serviceWorker.register("/static/js/sw.js");
      }
    </script>
    {{"{{"}} end {{"}}"}}
  </body>
</html>
{{"{{"}} end {{"}}"}}`

const tplLayout = tplLayoutTitleDesc + tplLayoutBaseOpen + tplLayoutNavFull + tplLayoutBaseClose

const tplLayoutMinimal = tplLayoutTitleDesc + tplLayoutBaseOpen + tplLayoutNavEmpty + tplLayoutBaseClose

const tplLayoutBlank = tplLayoutMinimal

const tplLayoutWelcome = `{{"{{"}} define "title" {{"}}"}}{{"{{"}} if .AppName {{"}}"}}{{"{{"}} .AppName {{"}}"}}{{"{{"}} else {{"}}"}}Cais{{"{{"}} end {{"}}"}}{{"{{"}} end {{"}}"}}
{{"{{"}} define "description" {{"}}"}}{{"{{"}} if .AppName {{"}}"}}{{"{{"}} .AppName {{"}}"}}{{"{{"}} else {{"}}"}}Cais{{"{{"}} end {{"}}"}} — powered by Cais{{"{{"}} end {{"}}"}}
{{"{{"}} define "welcome" {{"}}"}}
<!doctype html>
<html lang="{{"{{"}} htmlLang {{"}}"}}">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0, viewport-fit=cover" />
    {{"{{"}} if .CSRFToken {{"}}"}}<meta name="csrf-token" content="{{"{{"}} .CSRFToken {{"}}"}}" />{{"{{"}} end {{"}}"}}
    <title>{{"{{"}} template "title" . {{"}}"}}</title>
    <meta name="description" content="{{"{{"}} template "description" . {{"}}"}}" />
    <meta property="og:type" content="website" />
    <meta property="og:site_name" content="{{"{{"}} if .AppName {{"}}"}}{{"{{"}} .AppName {{"}}"}}{{"{{"}} else {{"}}"}}Cais{{"{{"}} end {{"}}"}}" />
    <meta property="og:title" content="{{"{{"}} template "title" . {{"}}"}}" />
    <meta property="og:description" content="{{"{{"}} template "description" . {{"}}"}}" />
    <meta property="og:image" content="{{"{{"}} absURL .AppURL "/static/og.png" {{"}}"}}" />
    <meta property="og:locale" content="{{"{{"}} ogLocale {{"}}"}}" />
    <meta name="twitter:card" content="summary_large_image" />
    <meta name="twitter:title" content="{{"{{"}} template "title" . {{"}}"}}" />
    <meta name="twitter:description" content="{{"{{"}} template "description" . {{"}}"}}" />
    <meta name="twitter:image" content="{{"{{"}} absURL .AppURL "/static/og.png" {{"}}"}}" />
    <link rel="stylesheet" href="/static/css/styles.css" />
    <link rel="manifest" href="/static/manifest.webmanifest" />
    <meta name="theme-color" content="#D4A574" />
    <meta name="mobile-web-app-capable" content="yes" />
    <meta name="apple-mobile-web-app-capable" content="yes" />
    <meta name="apple-mobile-web-app-status-bar-style" content="default" />
    <meta name="apple-mobile-web-app-title" content="{{"{{"}} if .AppName {{"}}"}}{{"{{"}} .AppName {{"}}"}}{{"{{"}} else {{"}}"}}Cais{{"{{"}} end {{"}}"}}" />
    <link rel="apple-touch-icon" href="/static/icons/icon.png" />
    <link rel="icon" href="/static/icons/icon.png" type="image/png" />
    <script src="/static/js/htmx.min.js" defer></script>
    <script src="/static/js/idiomorph-ext.min.js" defer></script>
    <script src="/static/js/sse-ext.min.js" defer></script>
    <script src="/static/js/cais.js" defer></script>
  </head>
  <body hx-ext="morph,sse" class="min-h-screen bg-gradient-to-b from-[#FAF3E8] via-[#EDCFA8] to-[#C9895E] text-stone-800 antialiased">
    <main>{{"{{"}} template "content" . {{"}}"}}</main>
    {{"{{"}} if eq .Env "development" {{"}}"}}
    <script>
      if ("serviceWorker" in navigator) {
        navigator.serviceWorker.getRegistrations().then(function (regs) {
          regs.forEach(function (r) { r.unregister(); });
        });
        if ("caches" in window) {
          caches.keys().then(function (keys) {
            keys.forEach(function (k) { caches.delete(k); });
          });
        }
      }
    </script>
    {{"{{"}} else {{"}}"}}
    <script>
      if ("serviceWorker" in navigator) {
        navigator.serviceWorker.register("/static/js/sw.js");
      }
    </script>
    {{"{{"}} end {{"}}"}}
  </body>
</html>
{{"{{"}} end {{"}}"}}
`

const tplCaisLogo = `{{"{{"}} define "cais_logo" {{"}}"}}
<img
  src="/static/img/go-on-cais.jpg"
  alt="Go on Cais"
  width="1024"
  height="683"
  class="w-full max-w-lg rounded-2xl shadow-xl shadow-amber-950/15 ring-1 ring-amber-900/10"
/>
{{"{{"}} end {{"}}"}}
`

const tplPageHome = `{{"{{"}} define "title" {{"}}"}}{{"{{"}} .AppName {{"}}"}}{{"{{"}} end {{"}}"}} {{"{{"}} define "content" {{"}}"}}
<div class="flex min-h-screen flex-col items-center justify-center px-6 py-14 text-center">
  {{"{{"}} template "cais_logo" . {{"}}"}}
  <h1 class="mt-10 font-serif text-4xl font-semibold tracking-tight text-stone-800 md:text-5xl">{{"{{"}} t "home.rails_heading" {{"}}"}}</h1>
  <p class="mt-3 max-w-md text-lg text-stone-600">{{"{{"}} t "home.rails_subtitle" .AppName {{"}}"}}</p>
  <p class="mt-6 text-sm font-medium uppercase tracking-[0.2em] text-amber-900/60">{{"{{"}} t "home.stack" {{"}}"}}</p>
  <div class="mt-12 w-full max-w-lg rounded-2xl border border-amber-900/10 bg-white/45 p-8 text-left shadow-xl shadow-amber-950/5 backdrop-blur-sm">
    <h2 class="mb-5 text-xs font-semibold uppercase tracking-wider text-stone-500">{{"{{"}} t "home.next_steps" {{"}}"}}</h2>
    <ol class="space-y-5 text-stone-700">
      <li class="flex gap-3">
        <span class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-amber-800/10 text-xs font-bold text-amber-950">1</span>
        <div>
          <p class="font-medium text-stone-800">{{"{{"}} t "home.step_resource" {{"}}"}}</p>
          <code class="mt-1.5 block rounded-lg bg-stone-100/90 px-3 py-2 font-mono text-xs text-stone-600">cais g resource item --fields name:string --public</code>
        </div>
      </li>
      <li class="flex gap-3">
        <span class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-amber-800/10 text-xs font-bold text-amber-950">2</span>
        <div>
          <p class="font-medium text-stone-800">{{"{{"}} t "home.step_dev" {{"}}"}}</p>
          <code class="mt-1.5 block rounded-lg bg-stone-100/90 px-3 py-2 font-mono text-xs text-stone-600">cais dev</code>
        </div>
      </li>
      <li class="flex gap-3">
        <span class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-amber-800/10 text-xs font-bold text-amber-950">3</span>
        <div>
          <p class="font-medium text-stone-800">{{"{{"}} t "home.step_docs" {{"}}"}}</p>
          <a href="https://github.com/puppe1990/cais" class="mt-1 inline-block text-sm text-amber-900 underline decoration-amber-700/40 underline-offset-2 hover:decoration-amber-800">github.com/puppe1990/cais</a>
        </div>
      </li>
    </ol>
  </div>
  <p class="mt-10 text-xs text-stone-500/90">{{"{{"}} t "home.powered_by" {{"}}"}}</p>
</div>
{{"{{"}} end {{"}}"}}
`

const tplPageContact = `{{"{{"}} define "title" {{"}}"}}{{"{{"}} t "contact.title" {{"}}"}}{{"{{"}} end {{"}}"}} {{"{{"}} define "content" {{"}}"}}
<div class="w-full max-w-md mx-auto p-4 md:p-5 bg-white rounded-xl border border-slate-200 shadow-2xs">
  <div class="mb-4 pb-4 border-b border-slate-100">
    <h2 class="text-lg font-black tracking-tight text-slate-900 font-display">{{"{{"}} t "contact.heading" {{"}}"}}</h2>
    <p class="text-[11px] text-slate-500 font-medium mt-1">{{"{{"}} t "contact.title" {{"}}"}}</p>
  </div>
  <form
    id="contact-form"
    {{"{{"}} hxForm "/contact" "#form-errors" "#contact-spinner" {{"}}"}}
  >
    <div id="form-errors"></div>
    <label class="block mb-2 text-sm font-medium text-slate-700" for="name">{{"{{"}} t "contact.name_label" {{"}}"}}</label>
    <input
      class="w-full border border-slate-300 rounded-lg px-3 py-2 mb-4 focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none"
      type="text"
      id="name"
      name="name"
      required
    />
    <label class="block mb-2 text-sm font-medium text-slate-700" for="email">{{"{{"}} t "contact.email_label" {{"}}"}}</label>
    <input
      class="w-full border border-slate-300 rounded-lg px-3 py-2 mb-4 focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none"
      type="email"
      id="email"
      name="email"
      required
    />
    <button
      class="w-full bg-slate-900 hover:bg-slate-800 text-white text-xs font-bold py-2.5 px-4 rounded-lg transition shadow-2xs"
      type="submit"
    >
      <span class="htmx-indicator" id="contact-spinner">{{"{{"}} t "contact.sending" {{"}}"}}</span>
      <span class="htmx-request-hide">{{"{{"}} t "contact.submit" {{"}}"}}</span>
    </button>
  </form>
</div>
{{"{{"}} end {{"}}"}}
`

const tplPageDashboard = `{{"{{"}} define "title" {{"}}"}}Dashboard{{"{{"}} end {{"}}"}} {{"{{"}} define "content" {{"}}"}}
<div class="w-full max-w-4xl mx-auto p-4 md:p-5 bg-white rounded-xl border border-slate-200 shadow-2xs">
  <div class="mb-4 pb-4 border-b border-slate-100">
    <h2 class="text-lg font-black tracking-tight text-slate-900 font-display">Dashboard</h2>
    <p class="text-[11px] text-slate-500 font-medium">Visão geral do seu app {{.AppName}}</p>
  </div>
  <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
    <div class="bg-slate-50 rounded-xl border border-slate-200 p-4 hover:shadow-2xs transition">
      <div class="flex items-center justify-between">
        <p class="text-[10px] font-bold text-slate-500 uppercase tracking-wider">Total Contacts</p>
        <span class="inline-flex items-center justify-center w-9 h-9 rounded-lg bg-indigo-100 text-indigo-600">
          {{"{{"}} template "icon_users_md" . {{"}}"}}
        </span>
      </div>
      <p class="mt-3 text-3xl font-black text-indigo-600 font-mono">{{"{{"}} .TotalContacts {{"}}"}}</p>
      <p class="mt-1 text-[10px] text-slate-400 font-medium">contatos cadastrados</p>
    </div>
    <div class="bg-slate-50 rounded-xl border border-slate-200 p-4 hover:shadow-2xs transition">
      <div class="flex items-center justify-between">
        <p class="text-[10px] font-bold text-slate-500 uppercase tracking-wider">Environment</p>
        <span class="inline-flex items-center justify-center w-9 h-9 rounded-lg bg-emerald-100 text-emerald-600">
          {{"{{"}} template "icon_shield_md" . {{"}}"}}
        </span>
      </div>
      <p class="mt-3 text-3xl font-black text-emerald-600 capitalize font-mono">{{"{{"}} .Env {{"}}"}}</p>
      <p class="mt-1 text-[10px] text-slate-400 font-medium">ambiente atual</p>
    </div>
  </div>
</div>
{{"{{"}} end {{"}}"}}
`

const tplPartialErrors = `{{"{{- "}}define "contact_errors" -{{"}}"}}
<div class="text-red-600 text-sm mb-4">{{"{{"}} .Message {{"}}"}}</div>
{{"{{- "}}end -{{"}}"}}
`

const tplPartialSuccess = `{{"{{- "}}define "contact_success" -{{"}}"}}
<div class="text-green-600 text-sm mb-4">{{"{{"}} t "contact.success" {{"}}"}}</div>
{{"{{- "}}end -{{"}}"}}
`

const tplGenericPage = `{{"{{"}} define "title" {{"}}"}}{{.Title}}{{"{{"}} end {{"}}"}} {{"{{"}} define "content" {{"}}"}}
<div class="bg-white p-6 rounded-2xl shadow-sm border border-slate-200 max-w-md mx-auto mt-10">
  <h2 class="text-2xl font-bold text-slate-800 mb-2">{{.Title}}</h2>
  <p class="text-slate-600">{{.Title}} page — customize this template.</p>
</div>
{{"{{"}} end {{"}}"}}
`
