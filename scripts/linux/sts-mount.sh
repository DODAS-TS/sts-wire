#!/bin/bash

client_name=""
iam_url=""
rgw_url=""
role=""
audience=""
bucket=""
mountpoint=""
pw_file=""

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
  echo "  -p, --pw-file PW_FILE    OIDC password file"
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
    -p|--pw-file)
      pw_file="$2"
      shift 2
      ;;
    -c|--client-name)
      client_name="$2"
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

if [ -z "$pw_file" ] || [ -z "$client_name" ] || [ -z "$iam_url" ] || [ -z "$rgw_url" ] || [ -z "$role" ] || [ -z "$audience" ] || [ -z "$bucket" ] || [ -z "$mountpoint" ]; then
  echo "Error: All options must be provided."
  usage
fi

oidc_account_list=$(oidc-add -l)
if ! echo "$oidc_account_list" | grep -q "^$client_name$"; then
  echo "Error: OIDC client '$client_name' does not exist in the list of client configurations."
  exit 1
fi

if [ ! -f "$pw_file" ]; then
  echo "Error: The specified pw-file '$pw_file' does not exist."
  exit 1
fi

if [ ! -d "$mountpoint" ]; then
  echo "Error: The specified mount point '$mountpoint' does not exist."
  exit 1
fi

if ! command -v sts-wire &> /dev/null; then
  echo "Error: sts-wire is not installed or not in the system's PATH."
  exit 1
fi

source oidc-agent-init.sh $client_name $audience $pw_file
export IAM_CLIENT_ID=$(oidc-gen -p $client_name --pw-file="$pw_file" | jq -r ".client_id")
export IAM_CLIENT_SECRET=$(oidc-gen -p $client_name --pw-file="$pw_file" | jq -r ".client_secret")
export REFRESH_TOKEN=$(oidc-gen -p $client_name --pw-file="$pw_file" | jq -r ".refresh_token")
wait_for_token
export ACCESS_TOKEN=$(cat /tmp/token)
sleep 5
sts-wire "$iam_url" myRGW "$rgw_url" "$role" "$audience" "$bucket" "$mountpoint" --tryRemount --noDummyFileCheck
