# awsw

AWS profile switcher for bash and zsh.

## Requirements

- Go 1.26 or later for development
- AWS CLI
- An AWS profile configured in `~/.aws/config` or `~/.aws/credentials`

## Build and install locally

```sh
go build -o awsw ./cmd/awsw
mkdir -p ~/.local/bin
cp awsw ~/.local/bin/awsw
```

Make sure `~/.local/bin` is in your `PATH`.

## Shell integration

Add the following line to `~/.zshrc` or `~/.bashrc`:

```sh
source /absolute/path/to/awsw/awsw.sh
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
