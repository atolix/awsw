package main

import (
	"errors"
	"fmt"
	"io"
	"os"
)

var errCancelled = errors.New("selection cancelled")

func main() {
	if len(os.Args) >= 2 && os.Args[1] == "init" {
		if len(os.Args) != 3 {
			fmt.Fprintln(os.Stderr, "usage: awsw init <zsh|bash>")
			os.Exit(2)
		}
		if err := printShellIntegration(os.Stdout, os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "awsw: %v\n", err)
			os.Exit(2)
		}
		return
	}

	if err := run(os.Stdin, os.Stdout, os.Stderr); err != nil {
		if !errors.Is(err, errCancelled) {
			fmt.Fprintf(os.Stderr, "awsw: %v\n", err)
		}
		if errors.Is(err, errCancelled) {
			os.Exit(0)
		}
		os.Exit(1)
	}
}

func printShellIntegration(out io.Writer, shell string) error {
	if shell != "zsh" && shell != "bash" {
		return fmt.Errorf("unsupported shell %q; use zsh or bash", shell)
	}

	_, err := fmt.Fprint(out, `# awsw shell integration
awsw() {
  if [ "$1" = "--clear" ] || [ "$1" = "-c" ]; then
    unset AWS_PROFILE
    unset AWS_DEFAULT_PROFILE
    echo "AWS profile cleared"
    return 0
  fi

  local profile
  profile="$(command awsw "$@")" || return $?

  if [ -n "$profile" ]; then
    export AWS_PROFILE="$profile"
    export AWS_DEFAULT_PROFILE="$profile"
    echo "AWS profile: $profile"
  fi
}
`)
	return err
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

	if err := ensureAuthenticated(profile, errOut); err != nil {
		return err
	}

	// stdout is intentionally reserved for the shell integration.
	_, err = fmt.Fprintln(out, profile)
	return err
}
