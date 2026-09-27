package main

import (
	"errors"
	"fmt"
	"io"
	"os"
)

var errCancelled = errors.New("selection cancelled")

const usageText = `usage: awsw [command]

  awsw                  select an AWS profile
  awsw --clear, -c      clear the current profile (requires the shell integration)
  awsw init <zsh|bash>  print the shell integration
  awsw --help, -h       show this help
`

func main() {
	os.Exit(execute(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

// execute dispatches on the command line arguments and returns the process
// exit code. Only the selected profile name is ever written to out; every
// other message goes to errOut so the shell integration, which captures
// stdout, never exports anything else as AWS_PROFILE.
func execute(args []string, in io.Reader, out, errOut io.Writer) int {
	if len(args) == 0 {
		return exitCode(run(in, out, errOut), errOut)
	}

	switch args[0] {
	case "init":
		if len(args) != 2 {
			fmt.Fprintln(errOut, "usage: awsw init <zsh|bash>")
			return 2
		}
		if err := printShellIntegration(out, args[1]); err != nil {
			fmt.Fprintf(errOut, "awsw: %v\n", err)
			return 2
		}
		return 0
	case "--help", "-h":
		fmt.Fprint(errOut, usageText)
		return 0
	case "--clear", "-c":
		fmt.Fprintln(errOut, `awsw: --clear needs the shell integration; add eval "$(awsw init zsh)" (or bash) to your shell configuration`)
		return 2
	default:
		fmt.Fprintf(errOut, "awsw: unknown argument %q\n\n%s", args[0], usageText)
		return 2
	}
}

func exitCode(err error, errOut io.Writer) int {
	if err == nil {
		return 0
	}
	if errors.Is(err, errCancelled) {
		return 0
	}
	fmt.Fprintf(errOut, "awsw: %v\n", err)
	return 1
}

func run(in io.Reader, out, errOut io.Writer) error {
	profiles, err := listProfiles()
	if err != nil {
		return err
	}
	if len(profiles) == 0 {
		return errors.New("no AWS profiles found")
	}

	profile, err := selectProfile(in, errOut, profiles)
	if err != nil {
		return err
	}

	if err := ensureAuthenticated(findProfile(profiles, profile), errOut); err != nil {
		return err
	}

	// stdout is intentionally reserved for the shell integration.
	_, err = fmt.Fprintln(out, profile)
	return err
}

func findProfile(profiles []profileDetails, name string) profileDetails {
	for _, profile := range profiles {
		if profile.Name == name {
			return profile
		}
	}
	return profileDetails{Name: name}
}
