package cli

import "fmt"

type resourceOpts struct {
	Fields    string
	Public    bool
	Seed      bool
	Paginate  bool
	AdminAuth string
	Force     bool
	dryRun    bool
}

func parseResourceOpts(args []string) (resourceOpts, error) {
	opts := resourceOpts{Seed: true, AdminAuth: "session"}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--fields":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--fields requires a value")
			}
			i++
			opts.Fields = args[i]
		case "--public":
			opts.Public = true
		case "--paginate":
			opts.Paginate = true
		case "--no-seed":
			opts.Seed = false
		case "--force":
			opts.Force = true
		case "--admin-auth":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--admin-auth requires a value")
			}
			i++
			switch args[i] {
			case "session", "bearer":
				opts.AdminAuth = args[i]
			default:
				return opts, fmt.Errorf("--admin-auth must be session or bearer")
			}
		default:
			return opts, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	return opts, nil
}
