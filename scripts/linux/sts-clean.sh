#!/bin/bash
for pid in $(pgrep sts-mount.sh); do kill -9 $pid; done
for pid in $(pgrep oidc-agent); do kill -9 $pid; done
for pid in $(pgrep sleep); do kill -9 $pid; done
rm -rf .myRGW/ /tmp/token
