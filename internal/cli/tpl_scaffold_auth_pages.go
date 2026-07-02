// Auth page templates for cais g auth and cais new (full app).
package cli

const tplPageLogin = `{{"{{"}} define "title" {{"}}"}}Login{{"{{"}} end {{"}}"}} {{"{{"}} define "content" {{"}}"}}
<div class="bg-white p-6 rounded-2xl shadow-sm border border-slate-200 max-w-md mx-auto mt-10">
  <h2 class="text-2xl font-bold text-slate-800 mb-4">{{"{{"}} t "auth.login_title" {{"}}"}}</h2>
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
    <button class="w-full bg-indigo-600 hover:bg-indigo-700 text-white font-medium py-2 px-4 rounded-xl transition" type="submit">
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
<div class="bg-white p-6 rounded-2xl shadow-sm border border-slate-200 max-w-md mx-auto mt-10">
  <h2 class="text-2xl font-bold text-slate-800 mb-4">{{"{{"}} t "auth.signup_title" {{"}}"}}</h2>
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
    <button class="w-full bg-indigo-600 hover:bg-indigo-700 text-white font-medium py-2 px-4 rounded-xl transition" type="submit">
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
<div class="bg-white p-6 rounded-2xl shadow-sm border border-slate-200 max-w-md mx-auto mt-10">
  <h2 class="text-2xl font-bold text-slate-800 mb-4">{{"{{"}} t "auth.forgot_password_title" {{"}}"}}</h2>
  <p class="text-sm text-slate-600 mb-4">{{"{{"}} t "auth.forgot_password_help" {{"}}"}}</p>
  <form method="post" action="/forgot-password" class="space-y-4">
    <input type="hidden" name="csrf_token" value="{{"{{"}} .CSRFToken {{"}}"}}" />
    <div>
      <label class="block text-sm font-medium text-slate-700 mb-1" for="email">{{"{{"}} t "contact.email_label" {{"}}"}}</label>
      <input class="w-full border border-slate-300 rounded-lg px-3 py-2" type="email" id="email" name="email" value="{{"{{"}} .Email {{"}}"}}" required />
      {{"{{"}} fieldError .Errors "email" {{"}}"}}
    </div>
    <button class="w-full bg-indigo-600 hover:bg-indigo-700 text-white font-medium py-2 px-4 rounded-xl transition" type="submit">
      {{"{{"}} t "auth.forgot_password_submit" {{"}}"}}
    </button>
  </form>
  <p class="text-sm text-slate-600 mt-4 text-center">
    <a class="text-indigo-600 hover:text-indigo-800" href="/login">{{"{{"}} t "auth.login_title" {{"}}"}}</a>
  </p>
</div>
{{"{{"}} end {{"}}"}}`

const tplPageResetPassword = `{{"{{"}} define "title" {{"}}"}}{{"{{"}} t "auth.reset_password_title" {{"}}"}}{{"{{"}} end {{"}}"}} {{"{{"}} define "content" {{"}}"}}
<div class="bg-white p-6 rounded-2xl shadow-sm border border-slate-200 max-w-md mx-auto mt-10">
  <h2 class="text-2xl font-bold text-slate-800 mb-4">{{"{{"}} t "auth.reset_password_title" {{"}}"}}</h2>
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
    <button class="w-full bg-indigo-600 hover:bg-indigo-700 text-white font-medium py-2 px-4 rounded-xl transition" type="submit">
      {{"{{"}} t "auth.reset_password_submit" {{"}}"}}
    </button>
  </form>
  {{"{{"}} end {{"}}"}}
  <p class="text-sm text-slate-600 mt-4 text-center">
    <a class="text-indigo-600 hover:text-indigo-800" href="/login">{{"{{"}} t "auth.login_title" {{"}}"}}</a>
  </p>
</div>
{{"{{"}} end {{"}}"}}`
