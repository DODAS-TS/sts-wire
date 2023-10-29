issuer=""
client_name=""

usage() {
  echo "Usage: $0 -i ISSUER -c CLIENT_NAME"
  echo "  -i, --issuer       Specify the issuer URL (e.g., https://iam.example.com)"
  echo "  -c, --client-name  Specify the client name (e.g., iam-client)"
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
    -h|--help)
      usage
      ;;
    *)
      echo "Error: Invalid option or argument."
      usage
      ;;
  esac
done

if [ -z "$issuer" ] || [ -z "$client_name" ]; then
  echo "Error: Both issuer and client name must be provided."
  usage
fi

eval $(oidc-agent-service use)
oidc-gen --pw-file=pw-file --scope-all --confirm-default --iss="$issuer" --flow=device "$client_name"
