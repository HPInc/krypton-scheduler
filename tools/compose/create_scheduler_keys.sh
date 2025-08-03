#!/bin/sh

rm -rf privateKey.env
if [ ! -f privateKey.pem ]; then
    openssl genrsa -out privateKey.pem 4096 
    openssl rsa -in privateKey.pem -pubout -out publicKey.pem
fi

echo "SCHEDULER_APP_PRIVATE_KEY_ENV=\"$(cat ./privateKey.pem)"\" > privateKey.env