package cli

import "fmt"

type appScaffoldOpts struct {
	dryRun bool
	data   bool
	force  bool
}

func parseAppOpts(args []string) (appScaffoldOpts, error) {
	opts := appScaffoldOpts{}
	for _, arg := range args {
		switch arg {
		case "--data":
			opts.data = true
		case "--force":
			opts.force = true
		default:
			return opts, fmt.Errorf("unknown app option %q (supported: --data, --force)", arg)
		}
	}
	return opts, nil
}
