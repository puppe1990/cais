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
    <script src="/static/js/cais.js" defer></script>
  </head>
  <body class="bg-slate-50 text-slate-900 min-h-screen flex flex-col">
    <header class="bg-white border-b border-slate-200 p-4 shadow-sm">
      <div class="max-w-5xl mx-auto flex justify-between items-center">
        <a href="/" class="font-bold text-xl text-indigo-600 hover:text-indigo-700 transition">{{.AppName}}</a>
        <nav class="flex items-center gap-6 text-sm font-medium">
          `

const tplLayoutNavFull = `<!-- cais:nav -->
          <a href="/" class="text-slate-600 hover:text-indigo-600 transition">Home</a>
          <a href="/contact" class="text-slate-600 hover:text-indigo-600 transition">Contact</a>
          <a href="/dashboard" class="text-slate-600 hover:text-indigo-600 transition">Dashboard</a>`

const tplLayoutNavEmpty = `<!-- cais:nav -->`

const tplLayoutBaseClose = `
        </nav>
      </div>
    </header>
    <main class="flex-grow max-w-5xl w-full mx-auto p-6">{{"{{"}} template "content" . {{"}}"}}</main>
    <footer class="border-t border-slate-200 p-4 text-center text-sm text-slate-500">
      {{.AppName}} — powered by Cais
    </footer>
    <script>
      if ("serviceWorker" in navigator) {
        navigator.serviceWorker.register("/static/js/sw.js");
      }
    </script>
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
    <script src="/static/js/cais.js" defer></script>
  </head>
  <body class="min-h-screen bg-gradient-to-b from-[#FAF3E8] via-[#EDCFA8] to-[#C9895E] text-stone-800 antialiased">
    <main>{{"{{"}} template "content" . {{"}}"}}</main>
    <script>
      if ("serviceWorker" in navigator) {
        navigator.serviceWorker.register("/static/js/sw.js");
      }
    </script>
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
<div class="bg-white p-6 rounded-2xl shadow-sm border border-slate-200 max-w-md mx-auto mt-10">
  <h2 class="text-2xl font-bold text-slate-800 mb-4">{{"{{"}} t "contact.heading" {{"}}"}}</h2>
  <form
    id="contact-form"
    hx-post="/contact"
    hx-target="#form-errors"
    hx-swap="innerHTML swap:150ms"
    hx-indicator="#contact-spinner"
    hx-disabled-elt="button[type='submit']"
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
      class="w-full bg-indigo-600 hover:bg-indigo-700 text-white font-medium py-2 px-4 rounded-xl transition"
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
<div class="space-y-8">
  <div>
    <h2 class="text-3xl font-bold text-slate-800">Dashboard</h2>
    <p class="text-slate-500 mt-1">Visão geral do seu app {{.AppName}}</p>
  </div>
  <div class="grid grid-cols-1 sm:grid-cols-2 gap-6">
    <div class="bg-white rounded-2xl shadow-sm border border-slate-200 p-6 hover:shadow-md transition">
      <div class="flex items-center justify-between">
        <p class="text-sm font-semibold text-slate-500 uppercase tracking-wide">Total Contacts</p>
        <span class="inline-flex items-center justify-center w-10 h-10 rounded-xl bg-indigo-100 text-indigo-600">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
        </span>
      </div>
      <p class="mt-4 text-4xl font-bold text-indigo-600">{{"{{"}} .TotalContacts {{"}}"}}</p>
      <p class="mt-1 text-sm text-slate-400">contatos cadastrados</p>
    </div>
    <div class="bg-white rounded-2xl shadow-sm border border-slate-200 p-6 hover:shadow-md transition">
      <div class="flex items-center justify-between">
        <p class="text-sm font-semibold text-slate-500 uppercase tracking-wide">Environment</p>
        <span class="inline-flex items-center justify-center w-10 h-10 rounded-xl bg-emerald-100 text-emerald-600">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
          </svg>
        </span>
      </div>
      <p class="mt-4 text-4xl font-bold text-emerald-600 capitalize">{{"{{"}} .Env {{"}}"}}</p>
      <p class="mt-1 text-sm text-slate-400">ambiente atual</p>
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
