export TMPDIR="/tmp"
export OIDC_CLIENT_NAME=$1
export AUDIENCE=$2
export DYLD_LIBRARY_PATH=$PWD/oidc-agent-3.3.5/lib/api

if [ ! -d oidc-agent-3.3.5 ]; then
    wget "https://github.com/indigo-dc/oidc-agent/archive/refs/tags/v3.3.5.tar.gz"
    tar xf v3.3.5.tar.gz
    rm -f v3.3.5.tar.gz
    cd oidc-agent-3.3.5
    make
    cd ..
fi

if [ ! -d jwt-tools ]; then
    git clone https://github.com/federicaagostini/useful-jwt-stuff.git ./jwt-tools
    pip3 install --user pyopenssl==22.0.0
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

eval $(./oidc-agent-3.3.5/bin/oidc-agent)

while true; do
    ./oidc-agent-3.3.5/bin/oidc-add --pw-cmd "cat pw-file" $OIDC_CLIENT_NAME
    if [ $? -eq 0 ]; then
        break
    else
        sleep 1
    fi
done

while true; do
    while ! check_token; do
        ./oidc-agent-3.3.5/bin/oidc-token $OIDC_CLIENT_NAME --time 1200 --aud=$AUDIENCE >"$TMPDIR/token"
        export ACCESS_TOKEN=$(cat "$TMPDIR/token")
    done
    sleep 600
done &
