package cli

type buildOptions struct {
	GOOS   string
	GOARCH string
	Output string
}

func parseBuildFlags(args []string) buildOptions {
	opts := buildOptions{Output: serverBin}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--os":
			if i+1 < len(args) {
				i++
				opts.GOOS = args[i]
			}
		case "--arch":
			if i+1 < len(args) {
				i++
				opts.GOARCH = args[i]
			}
		case "-o", "--output":
			if i+1 < len(args) {
				i++
				opts.Output = args[i]
			}
		}
	}
	return opts
}
