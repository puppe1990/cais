package cli

const tplI18nCatalog = `package i18n

import (
	caisi18n "github.com/puppe1990/cais/pkg/cais/i18n"
)

var locales = map[string]map[string]string{
	"en": enMessages,
	"pt": ptMessages,
}

// NewCatalog returns a catalog for the given locale (en default, pt for pt-BR).
func NewCatalog(locale string) *caisi18n.Catalog {
	return caisi18n.NewCatalogFrom(locale, locales)
}

// DefaultCatalog returns the English catalog.
func DefaultCatalog() *caisi18n.Catalog {
	return NewCatalog(caisi18n.DefaultLocale)
}
`

const tplI18nEn = `package i18n

var enMessages = map[string]string{
	"auth.invalid_credentials": "Invalid email or password.",
	"auth.welcome":             "Welcome!",
	"auth.login_title":         "Sign in",
	"auth.login_submit":        "Sign in",
	"auth.password_label":              "Password",
	"auth.password_confirmation_label": "Confirm password",
	"auth.logout":                      "Sign out",
	"auth.forgot_password":             "Forgot password?",
	"auth.forgot_password_title":       "Reset your password",
	"auth.forgot_password_help":        "Enter your email and we'll send reset instructions.",
	"auth.forgot_password_submit":      "Send reset link",
	"auth.reset_password_title":        "Choose a new password",
	"auth.reset_password_submit":       "Update password",
	"auth.reset_email_sent":            "If that email is registered, you will receive reset instructions shortly.",
	"auth.reset_success":               "Your password was updated. Sign in with your new password.",
	"auth.reset_invalid_token":         "This reset link is invalid or has expired.",
	"auth.password_too_short":          "Password must be at least 8 characters.",
	"auth.password_mismatch":           "Passwords do not match.",
	"auth.signup_title":                "Create account",
	"auth.signup_submit":               "Sign up",
	"auth.signup_prompt":               "Don't have an account?",
	"auth.login_prompt":                "Already have an account?",
	"auth.email_taken":                 "This email is already registered.",

	"contact.title":          "Contact",
	"contact.heading":        "Get in touch",
	"contact.name_label":     "Name",
	"contact.name_required":  "Name is required.",
	"contact.email_label":    "Email",
	"contact.email_required": "Email is required.",
	"contact.email_invalid":  "Enter a valid email.",
	"contact.submit":         "Send",
	"contact.sending":        "Sending…",
	"contact.success":        "Message sent successfully!",

	"home.title":            "Home",
	"home.welcome":          "Welcome, %s!",
	"home.tagline":          "Mini Go app with HTMX, Tailwind, and SQLite.",
	"home.contact_link":     "Contact",
	"home.default_name":     "Developer",
	"home.rails_heading":    "You're on Cais!",
	"home.rails_subtitle":   "%s is ready to sail.",
	"home.stack":            "Go · HTMX · Tailwind · SQLite",
	"home.next_steps":       "Next steps",
	"home.step_resource":    "Generate your first resource",
	"home.step_dev":         "Start the dev server",
	"home.step_docs":        "Explore the framework",
	"home.powered_by":       "Powered by Cais — lightweight apps on Lightsail",
	"home.minimal.tagline":  "Go app with HTMX, Tailwind, and SQLite — powered by Cais.",
	"home.minimal.hint":     "Use ` + "`cais g resource <name> --public`" + ` to get started.",

	"dashboard.title":    "Dashboard",
	"dashboard.contacts": "Contacts:",
	"dashboard.env":      "Environment:",

	"layout.footer": "Running light on Lightsail",
}
`

const tplI18nPt = `package i18n

var ptMessages = map[string]string{
	"auth.invalid_credentials": "Email ou senha inválidos.",
	"auth.welcome":             "Bem-vindo!",
	"auth.login_title":         "Entrar",
	"auth.login_submit":        "Entrar",
	"auth.password_label":              "Senha",
	"auth.password_confirmation_label": "Confirmar senha",
	"auth.logout":                      "Sair",
	"auth.forgot_password":             "Esqueceu a senha?",
	"auth.forgot_password_title":       "Redefinir senha",
	"auth.forgot_password_help":        "Informe seu email para receber as instruções.",
	"auth.forgot_password_submit":      "Enviar link",
	"auth.reset_password_title":        "Escolha uma nova senha",
	"auth.reset_password_submit":       "Atualizar senha",
	"auth.reset_email_sent":            "Se esse email estiver cadastrado, você receberá as instruções em breve.",
	"auth.reset_success":               "Senha atualizada. Entre com a nova senha.",
	"auth.reset_invalid_token":         "Este link é inválido ou expirou.",
	"auth.password_too_short":          "A senha deve ter pelo menos 8 caracteres.",
	"auth.password_mismatch":           "As senhas não coincidem.",
	"auth.signup_title":                "Criar conta",
	"auth.signup_submit":               "Cadastrar",
	"auth.signup_prompt":               "Não tem conta?",
	"auth.login_prompt":                "Já tem conta?",
	"auth.email_taken":                 "Este email já está cadastrado.",

	"contact.title":          "Contato",
	"contact.heading":        "Fale conosco",
	"contact.name_label":     "Nome",
	"contact.name_required":  "O campo nome é obrigatório.",
	"contact.email_label":    "Email",
	"contact.email_required": "O campo email é obrigatório.",
	"contact.email_invalid":  "Informe um email válido.",
	"contact.submit":         "Enviar",
	"contact.sending":        "Enviando…",
	"contact.success":        "Mensagem enviada com sucesso!",

	"home.title":            "Página Inicial",
	"home.welcome":          "Bem-vindo, %s!",
	"home.tagline":          "Mini app Go com HTMX, Tailwind e SQLite.",
	"home.contact_link":     "Contato",
	"home.default_name":     "Desenvolvedor",
	"home.rails_heading":    "Você está no Cais!",
	"home.rails_subtitle":   "%s está pronto para navegar.",
	"home.stack":            "Go · HTMX · Tailwind · SQLite",
	"home.next_steps":       "Próximos passos",
	"home.step_resource":    "Gere seu primeiro resource",
	"home.step_dev":         "Suba o servidor de desenvolvimento",
	"home.step_docs":        "Explore o framework",
	"home.powered_by":       "Powered by Cais — apps leves no Lightsail",
	"home.minimal.tagline":  "App Go com HTMX, Tailwind e SQLite — powered by Cais.",
	"home.minimal.hint":     "Use ` + "`cais g resource <name> --public`" + ` para começar.",

	"dashboard.title":    "Dashboard",
	"dashboard.contacts": "Contatos:",
	"dashboard.env":      "Ambiente:",

	"layout.footer": "Rodando leve no Lightsail",
}
`

const tplSeeds = `package db

import (
	"{{.ModulePath}}/internal/models"
	"{{.ModulePath}}/internal/store"
)

// RunSeeds populates demo data. Safe to run multiple times.
func RunSeeds(s store.Store) error {
	// cais:recurring-seeds
	// cais:seeds
	if _, err := s.InsertContact(models.Contact{
		Name:  "Demo",
		Email: "demo@example.com",
	}); err != nil {
		return err
	}
	return nil
}
`

const tplSeedsMinimal = `package db

import "{{.ModulePath}}/internal/store"

// RunSeeds populates demo data. Safe to run multiple times.
func RunSeeds(s store.Store) error {
	// cais:recurring-seeds
	// cais:seeds
	return nil
}
`

const tplI18nTest = `package i18n

import "testing"

func TestDefaultCatalog_english(t *testing.T) {
	c := DefaultCatalog()
	if got := c.T("auth.welcome"); got != "Welcome!" {
		t.Errorf("T(auth.welcome) = %q", got)
	}
	if got := c.T("auth.signup_prompt"); got != "Don't have an account?" {
		t.Errorf("T(auth.signup_prompt) = %q", got)
	}
}

func TestNewCatalog_portuguese(t *testing.T) {
	c := NewCatalog("pt-BR")
	if got := c.T("auth.welcome"); got != "Bem-vindo!" {
		t.Errorf("T(auth.welcome) = %q", got)
	}
	if c.HTMLLang() != "pt-BR" {
		t.Errorf("HTMLLang() = %q, want pt-BR", c.HTMLLang())
	}
}
`
