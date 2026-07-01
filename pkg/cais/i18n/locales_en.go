package i18n

var enMessages = map[string]string{
	"auth.invalid_credentials": "Invalid email or password.",
	"auth.welcome":             "Welcome!",
	"auth.login_title":         "Sign in",
	"auth.login_submit":        "Sign in",
	"auth.password_label":      "Password",
	"auth.logout":              "Sign out",

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

	"home.title":          "Home",
	"home.welcome":        "Welcome, %s!",
	"home.tagline":        "Mini Go app with HTMX, Tailwind, and SQLite.",
	"home.contact_link":   "Contact",
	"home.default_name":   "Developer",
	"home.rails_heading":  "You're on Cais!",
	"home.rails_subtitle": "%s is ready to sail.",
	"home.stack":          "Go · HTMX · Tailwind · SQLite",
	"home.next_steps":     "Next steps",
	"home.step_resource":  "Generate your first resource",
	"home.step_dev":       "Start the dev server",
	"home.step_docs":      "Explore the framework",
	"home.powered_by":     "Powered by Cais — lightweight apps on Lightsail",

	"dashboard.title":    "Dashboard",
	"dashboard.contacts": "Contacts:",
	"dashboard.env":      "Environment:",

	"layout.footer": "Running light on Lightsail",
}
