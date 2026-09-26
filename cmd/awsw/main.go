package main

import (
	"errors"
	"fmt"
	"io"
	"os"
)

var errCancelled = errors.New("selection cancelled")

func main() {
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
