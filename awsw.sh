# Shell integration for awsw-bin.
# Source this file from ~/.zshrc or ~/.bashrc.
awsw() {
  if [ "$1" = "--clear" ] || [ "$1" = "-c" ]; then
    unset AWS_PROFILE
    unset AWS_DEFAULT_PROFILE
    echo "AWS profile cleared"
    return 0
  fi

  local profile
  profile="$(awsw-bin "$@")" || return $?

  if [ -n "$profile" ]; then
    export AWS_PROFILE="$profile"
    export AWS_DEFAULT_PROFILE="$profile"
    echo "AWS profile: $profile"
  fi
}
