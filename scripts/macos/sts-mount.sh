OIDC_CLIENT_NAME=""
iam_url=""
rgw_url=""
role=""
audience=""
bucket=""
mountpoint=""

wait_for_token() {
    while true; do
        if [ -f /tmp/token ] && [ -s /tmp/token ]; then
            break
        else
            sleep 3
        fi
    done
}

usage() {
  echo "Usage: $0 [OPTIONS]"
  echo "  -c, --client-name NAME   OIDC client name"
  echo "  -i, --iam-url URL        IAM URL (e.g., https://iam.example.com/)"
  echo "  -r, --rgw-url URL        RGW URL (e.g., https://rgw.example.com/)"
  echo "  -o, --role ROLE          Role (e.g., S3AccessIAMRole)"
  echo "  -a, --audience AUDIENCE  Audience (e.g., https://wlcg.cern.ch/jwt/v1/any)"
  echo "  -b, --bucket BUCKET      Bucket (e.g., /bucket)"
  echo "  -m, --mountpoint PATH    Mountpoint (e.g., ./mountpoint/)"
  echo "  -h, --help               Display this help message"
  exit 1
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    -c|--client-name)
      OIDC_CLIENT_NAME="$2"
      shift 2
      ;;
    -i|--iam-url)
      iam_url="$2"
      shift 2
      ;;
    -r|--rgw-url)
      rgw_url="$2"
      shift 2
      ;;
    -o|--role)
      role="$2"
      shift 2
      ;;
    -a|--audience)
      audience="$2"
      shift 2
      ;;
    -b|--bucket)
      bucket="$2"
      shift 2
      ;;
    -m|--mountpoint)
      mountpoint="$2"
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

if [ -z "$OIDC_CLIENT_NAME" ] || [ -z "$iam_url" ] || [ -z "$rgw_url" ] || [ -z "$role" ] || [ -z "$audience" ] || [ -z "$bucket" ] || [ -z "$mountpoint" ]; then
  echo "Error: All options must be provided."
  usage
fi

oidc_account_list=$(oidc-add -l)
if ! echo "$oidc_account_list" | grep -q "^$OIDC_CLIENT_NAME$"; then
  echo "Error: OIDC client '$OIDC_CLIENT_NAME' does not exist in the list of client configurations."
  exit 1
fi

if [ ! -x ./sts-wire_osx ]; then
  echo "Error: sts-wire_osx binary does not exist or is not executable."
  exit 1
fi

source oidc-agent-init.sh $OIDC_CLIENT_NAME $audience
export IAM_CLIENT_ID=$(oidc-gen -p $OIDC_CLIENT_NAME --pw-file=pw-file | jq -r ".client_id")
export IAM_CLIENT_SECRET=$(oidc-gen -p $OIDC_CLIENT_NAME --pw-file=pw-file | jq -r ".client_secret")
export REFRESH_TOKEN=$(oidc-gen -p $OIDC_CLIENT_NAME --pw-file=pw-file | jq -r ".refresh_token")
wait_for_token
export ACCESS_TOKEN=$(cat /tmp/token)
sleep 5
./sts-wire_osx "$iam_url" myRGW "$rgw_url" "$role" "$audience" "$bucket" "$mountpoint" --tryRemount --noDummyFileCheck
