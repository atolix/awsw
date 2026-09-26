package main

import (
	"fmt"
	"io"
)

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
