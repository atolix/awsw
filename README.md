# awsw

AWS profile switcher for bash and zsh.

## Requirements

- Go 1.26 or later for development
- AWS CLI
- An AWS profile configured in `~/.aws/config` or `~/.aws/credentials`

## Install

```sh
go install github.com/atolix/awsw/cmd/awsw@latest
```

Make sure the Go binary install directory is in your `PATH`.

## Shell integration

Add one line to `~/.zshrc`:

```sh
eval "$(awsw init zsh)"
```

For bash, use:

```sh
eval "$(awsw init bash)"
```

Reload the shell configuration:

```sh
source ~/.zshrc
```

## Usage

Select a profile by number:

```sh
awsw
```

The selector supports arrow keys, `/` to filter profiles, `Enter` to select, and `Esc` or `Ctrl-C` to cancel.

If the selected profile is an AWS SSO profile and its session is not valid, `awsw` runs `aws sso login` before switching profiles.

Clear the current profile:

```sh
awsw --clear
```

The Go binary prints only the selected profile name to stdout. This allows the shell function to export `AWS_PROFILE` and `AWS_DEFAULT_PROFILE` in the current shell.
