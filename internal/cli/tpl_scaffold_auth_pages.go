// Auth page templates for cais g auth and cais new (full app).
package cli

const tplPageLogin = `{{"{{"}} define "title" {{"}}"}}Login{{"{{"}} end {{"}}"}} {{"{{"}} define "content" {{"}}"}}
<div class="w-full max-w-md mx-auto p-4 md:p-5 bg-white rounded-xl border border-slate-200 shadow-2xs">
  <div class="mb-4 pb-4 border-b border-slate-100">
    <h2 class="text-lg font-black tracking-tight text-slate-900 font-display">{{"{{"}} t "auth.login_title" {{"}}"}}</h2>
  </div>
  {{"{{"}} if .Error {{"}}"}}<p class="text-red-600 text-sm mb-4">{{"{{"}} .Error {{"}}"}}</p>{{"{{"}} end {{"}}"}}
  <form method="post" action="/login" class="space-y-4">
    <input type="hidden" name="csrf_token" value="{{"{{"}} .CSRFToken {{"}}"}}" />
    <div>
      <label class="block text-sm font-medium text-slate-700 mb-1" for="email">Email</label>
      <input class="w-full border border-slate-300 rounded-lg px-3 py-2" type="email" id="email" name="email" required />
    </div>
    <div>
      <label class="block text-sm font-medium text-slate-700 mb-1" for="password">{{"{{"}} t "auth.password_label" {{"}}"}}</label>
      <input class="w-full border border-slate-300 rounded-lg px-3 py-2" type="password" id="password" name="password" required />
    </div>
    <button class="w-full bg-slate-900 hover:bg-slate-800 text-white text-xs font-bold py-2.5 px-4 rounded-lg transition shadow-2xs" type="submit">
      {{"{{"}} t "auth.login_submit" {{"}}"}}
    </button>
  </form>
  <p class="text-sm text-slate-600 mt-4 text-center space-y-1">
    <span class="block">
      {{"{{"}} t "auth.signup_prompt" {{"}}"}}
      <a class="text-indigo-600 hover:text-indigo-800" href="/signup">{{"{{"}} t "auth.signup_title" {{"}}"}}</a>
    </span>
    <a class="text-indigo-600 hover:text-indigo-800" href="/forgot-password">{{"{{"}} t "auth.forgot_password" {{"}}"}}</a>
  </p>
</div>
{{"{{"}} end {{"}}"}}
`

const tplPageSignup = `{{"{{"}} define "title" {{"}}"}}{{"{{"}} t "auth.signup_title" {{"}}"}}{{"{{"}} end {{"}}"}} {{"{{"}} define "content" {{"}}"}}
<div class="w-full max-w-md mx-auto p-4 md:p-5 bg-white rounded-xl border border-slate-200 shadow-2xs">
  <div class="mb-4 pb-4 border-b border-slate-100">
    <h2 class="text-lg font-black tracking-tight text-slate-900 font-display">{{"{{"}} t "auth.signup_title" {{"}}"}}</h2>
  </div>
  <form method="post" action="/signup" class="space-y-4">
    <input type="hidden" name="csrf_token" value="{{"{{"}} .CSRFToken {{"}}"}}" />
    <div>
      <label class="block text-sm font-medium text-slate-700 mb-1" for="email">{{"{{"}} t "contact.email_label" {{"}}"}}</label>
      <input class="w-full border border-slate-300 rounded-lg px-3 py-2" type="email" id="email" name="email" value="{{"{{"}} .Email {{"}}"}}" required />
      {{"{{"}} fieldError .Errors "email" {{"}}"}}
    </div>
    <div>
      <label class="block text-sm font-medium text-slate-700 mb-1" for="password">{{"{{"}} t "auth.password_label" {{"}}"}}</label>
      <input class="w-full border border-slate-300 rounded-lg px-3 py-2" type="password" id="password" name="password" required />
      {{"{{"}} fieldError .Errors "password" {{"}}"}}
    </div>
    <div>
      <label class="block text-sm font-medium text-slate-700 mb-1" for="password_confirmation">{{"{{"}} t "auth.password_confirmation_label" {{"}}"}}</label>
      <input class="w-full border border-slate-300 rounded-lg px-3 py-2" type="password" id="password_confirmation" name="password_confirmation" required />
      {{"{{"}} fieldError .Errors "password_confirmation" {{"}}"}}
    </div>
    <button class="w-full bg-slate-900 hover:bg-slate-800 text-white text-xs font-bold py-2.5 px-4 rounded-lg transition shadow-2xs" type="submit">
      {{"{{"}} t "auth.signup_submit" {{"}}"}}
    </button>
  </form>
  <p class="text-sm text-slate-600 mt-4 text-center">
    {{"{{"}} t "auth.login_prompt" {{"}}"}}
    <a class="text-indigo-600 hover:text-indigo-800" href="/login">{{"{{"}} t "auth.login_title" {{"}}"}}</a>
  </p>
</div>
{{"{{"}} end {{"}}"}}`

const tplPageForgotPassword = `{{"{{"}} define "title" {{"}}"}}{{"{{"}} t "auth.forgot_password_title" {{"}}"}}{{"{{"}} end {{"}}"}} {{"{{"}} define "content" {{"}}"}}
<div class="w-full max-w-md mx-auto p-4 md:p-5 bg-white rounded-xl border border-slate-200 shadow-2xs">
  <div class="mb-4 pb-4 border-b border-slate-100">
    <h2 class="text-lg font-black tracking-tight text-slate-900 font-display">{{"{{"}} t "auth.forgot_password_title" {{"}}"}}</h2>
    <p class="text-[11px] text-slate-500 font-medium mt-1">{{"{{"}} t "auth.forgot_password_help" {{"}}"}}</p>
  </div>
  <form method="post" action="/forgot-password" class="space-y-4">
    <input type="hidden" name="csrf_token" value="{{"{{"}} .CSRFToken {{"}}"}}" />
    <div>
      <label class="block text-sm font-medium text-slate-700 mb-1" for="email">{{"{{"}} t "contact.email_label" {{"}}"}}</label>
      <input class="w-full border border-slate-300 rounded-lg px-3 py-2" type="email" id="email" name="email" value="{{"{{"}} .Email {{"}}"}}" required />
      {{"{{"}} fieldError .Errors "email" {{"}}"}}
    </div>
    <button class="w-full bg-slate-900 hover:bg-slate-800 text-white text-xs font-bold py-2.5 px-4 rounded-lg transition shadow-2xs" type="submit">
      {{"{{"}} t "auth.forgot_password_submit" {{"}}"}}
    </button>
  </form>
  <p class="text-sm text-slate-600 mt-4 text-center">
    <a class="text-indigo-600 hover:text-indigo-800" href="/login">{{"{{"}} t "auth.login_title" {{"}}"}}</a>
  </p>
</div>
{{"{{"}} end {{"}}"}}`

const tplPageResetPassword = `{{"{{"}} define "title" {{"}}"}}{{"{{"}} t "auth.reset_password_title" {{"}}"}}{{"{{"}} end {{"}}"}} {{"{{"}} define "content" {{"}}"}}
<div class="w-full max-w-md mx-auto p-4 md:p-5 bg-white rounded-xl border border-slate-200 shadow-2xs">
  <div class="mb-4 pb-4 border-b border-slate-100">
    <h2 class="text-lg font-black tracking-tight text-slate-900 font-display">{{"{{"}} t "auth.reset_password_title" {{"}}"}}</h2>
  </div>
  {{"{{"}} if .Error {{"}}"}}<p class="text-red-600 text-sm mb-4">{{"{{"}} .Error {{"}}"}}</p>{{"{{"}} end {{"}}"}}
  {{"{{"}} if .Token {{"}}"}}
  <form method="post" action="/reset-password" class="space-y-4">
    <input type="hidden" name="csrf_token" value="{{"{{"}} .CSRFToken {{"}}"}}" />
    <input type="hidden" name="token" value="{{"{{"}} .Token {{"}}"}}" />
    <div>
      <label class="block text-sm font-medium text-slate-700 mb-1" for="password">{{"{{"}} t "auth.password_label" {{"}}"}}</label>
      <input class="w-full border border-slate-300 rounded-lg px-3 py-2" type="password" id="password" name="password" required />
      {{"{{"}} fieldError .Errors "password" {{"}}"}}
    </div>
    <div>
      <label class="block text-sm font-medium text-slate-700 mb-1" for="password_confirmation">{{"{{"}} t "auth.password_confirmation_label" {{"}}"}}</label>
      <input class="w-full border border-slate-300 rounded-lg px-3 py-2" type="password" id="password_confirmation" name="password_confirmation" required />
      {{"{{"}} fieldError .Errors "password_confirmation" {{"}}"}}
    </div>
    <button class="w-full bg-slate-900 hover:bg-slate-800 text-white text-xs font-bold py-2.5 px-4 rounded-lg transition shadow-2xs" type="submit">
      {{"{{"}} t "auth.reset_password_submit" {{"}}"}}
    </button>
  </form>
  {{"{{"}} end {{"}}"}}
  <p class="text-sm text-slate-600 mt-4 text-center">
    <a class="text-indigo-600 hover:text-indigo-800" href="/login">{{"{{"}} t "auth.login_title" {{"}}"}}</a>
  </p>
</div>
{{"{{"}} end {{"}}"}}`
