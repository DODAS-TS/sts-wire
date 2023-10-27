#!/bin/bash

issuer=""
client_name=""
pw_file=""

usage() {
  echo "Usage: $0 -i ISSUER -c CLIENT_NAME -p PW_FILE"
  echo "  -i, --issuer       Specify the issuer URL (e.g., https://iam.example.com)"
  echo "  -c, --client-name  Specify the client name (e.g., iam-client)"
  echo "  -p, --pw-file      Specify the password file"
  echo "  -h, --help         Display this help message"
  exit 1
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    -i|--issuer)
      issuer="$2"
      shift 2
      ;;
    -c|--client-name)
      client_name="$2"
      shift 2
      ;;
    -p|--pw-file)
      pw_file="$2"
      shift 2
      ;;
    -h|--help)
      usage
      ;;
    *)
      echo "Error: Invalid option or argument."
      usage
      ;;
  esac
done

if [ -z "$issuer" ] || [ -z "$client_name" ] || [ -z "$pw_file" ]; then
  echo "Error: Issuer, client name, and pw-file must be provided."
  usage
fi

if [ ! -f "$pw_file" ]; then
  echo "Error: The specified pw-file '$pw_file' does not exist."
  exit 1
fi

mkdir -p $HOME/.oidc-agent
export OIDC_CONFIG_DIR=$HOME/.oidc-agent
export OIDC_AGENT=/usr/bin/oidc-agent
eval $(oidc-agent-service use)
oidc-gen --pw-file="$pw_file" --scope-all --confirm-default --iss="$issuer" --flow=device "$client_name"
