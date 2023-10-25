#!/bin/bash

export TMPDIR="/tmp"
export OIDC_AGENT=/usr/bin/oidc-agent
export OIDC_CLIENT_NAME=$1
export AUDIENCE=$2
export PW_FILE=$3

while true; do
    if [ ! -d jwt-tools ]; then
        git clone https://github.com/federicaagostini/useful-jwt-stuff.git ./jwt-tools
    fi
    if [ -d jwt-tools ]; then
        break
    fi
    sleep 1
done

if ! pip3 freeze | grep -q 'pyOpenSSL==22.0.0'; then
  pip3 install --user pyopenssl==22.0.0
fi

if ! pip3 freeze | grep -q -f ./jwt-tools/requirements.txt; then
  pip3 install --user -r ./jwt-tools/requirements.txt
fi

check_token() {
  if [ -s "$TMPDIR/token" ]; then
    local current_time=$(date +%s)
    local expiration_time=$(python3 ./jwt-tools/JWTdecode.py $(cat "$TMPDIR/token") | jq -r '.exp')
    if [ "$expiration_time" -gt "$current_time" ]; then
      return 0
    else
      return 1
    fi
  else
    return 2
  fi
}

eval $(oidc-keychain)

while true; do
    oidc-add --pw-file=$PW_FILE $OIDC_CLIENT_NAME
    if [ $? -eq 0 ]; then
        break
    else
        sleep 1
    fi
done

while true; do
    while ! check_token; do
        oidc-token $OIDC_CLIENT_NAME --time 1200 --aud=$AUDIENCE >"$TMPDIR/token"
        export ACCESS_TOKEN=$(cat "$TMPDIR/token")
    done
    sleep 600
done &
