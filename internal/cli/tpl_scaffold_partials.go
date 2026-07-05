package cli

const tplPartialIcons = `{{"{{"}}- define "icon_sparkles_md" {{"}}"}}-
<svg class="w-5 h-5 text-amber-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 3v4M3 5h4M6 17v4m-2-2h4m5-16l2.286 6.857L21 12l-5.714 2.143L13 21l-2.286-6.857L5 12l5.714-2.143L13 3z" /></svg>
{{"{{"}}- end {{"}}"}}-
{{"{{"}}- define "icon_home_nav" {{"}}"}}-
<svg class="cais-nav-icon" width="14" height="14" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" /></svg>
{{"{{"}}- end {{"}}"}}-
{{"{{"}}- define "icon_message_nav" {{"}}"}}-
<svg class="cais-nav-icon" width="14" height="14" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" /></svg>
{{"{{"}}- end {{"}}"}}-
{{"{{"}}- define "icon_chart_nav" {{"}}"}}-
<svg class="cais-nav-icon" width="14" height="14" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" /></svg>
{{"{{"}}- end {{"}}"}}-
{{"{{"}}- define "icon_users_md" {{"}}"}}-
<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" /></svg>
{{"{{"}}- end {{"}}"}}-
{{"{{"}}- define "icon_shield_md" {{"}}"}}-
<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" /></svg>
{{"{{"}}- end {{"}}"}}-
`

const tplPartialNavLinks = `{{"{{"}}- define "nav_links" {{"}}"}}-
<a href="/" data-cais-nav="/" hx-boost="true" hx-target="#cais-main" hx-select="#cais-main" hx-push-url="true" hx-swap="morph:innerHTML" class="px-3 py-1.5 rounded-lg text-xs font-bold transition-all flex items-center gap-1.5 flex-shrink-0 {{"{{"}} if eq .ActiveNav "home" {{"}}"}}bg-slate-900 text-white shadow-2xs{{"{{"}} else {{"}}"}}text-slate-600 hover:text-slate-900 hover:bg-slate-100{{"{{"}} end {{"}}"}}">{{"{{"}} template "icon_home_nav" . {{"}}"}}Home</a>
<a href="/contact" data-cais-nav="/contact" hx-boost="true" hx-target="#cais-main" hx-select="#cais-main" hx-push-url="true" hx-swap="morph:innerHTML" class="px-3 py-1.5 rounded-lg text-xs font-bold transition-all flex items-center gap-1.5 flex-shrink-0 {{"{{"}} if eq .ActiveNav "contact" {{"}}"}}bg-slate-900 text-white shadow-2xs{{"{{"}} else {{"}}"}}text-slate-600 hover:text-slate-900 hover:bg-slate-100{{"{{"}} end {{"}}"}}">{{"{{"}} template "icon_message_nav" . {{"}}"}}Contact</a>
<a href="/dashboard" data-cais-nav="/dashboard" hx-boost="true" hx-target="#cais-main" hx-select="#cais-main" hx-push-url="true" hx-swap="morph:innerHTML" class="px-3 py-1.5 rounded-lg text-xs font-bold transition-all flex items-center gap-1.5 flex-shrink-0 {{"{{"}} if eq .ActiveNav "dashboard" {{"}}"}}bg-slate-900 text-white shadow-2xs{{"{{"}} else {{"}}"}}text-slate-600 hover:text-slate-900 hover:bg-slate-100{{"{{"}} end {{"}}"}}">{{"{{"}} template "icon_chart_nav" . {{"}}"}}Dashboard</a>
{{"{{"}}- end {{"}}"}}-
`
