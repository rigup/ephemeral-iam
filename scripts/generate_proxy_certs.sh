#!/bin/bash

workdir="${GOPATH}/src/github.com/jessesomerville/gcp-iam-escalate"

if [[ $(/bin/ls -1 ${workdir}/server.* 2> /dev/null | wc -l) -lt 2 ]]; then
    openssl req -new -newkey rsa:4096 -nodes -x509 -days 365 \
        -subj "/C=US/ST=Unknown/L=Unknown/O=Unknown/CN=Unknown" \
        -keyout ${workdir}/server.key -out ${workdir}/server.pem \
        &> /dev/null
fi