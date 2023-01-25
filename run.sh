#!/bin/bash

if [ ! -f ./mkcert ]; then
	curl -JLO "https://dl.filippo.io/mkcert/latest?for=linux/amd64"
	chmod +x mkcert-v*-linux-amd64
	mv mkcert-v*-linux-amd64 mkcert
fi
  
./mkcert -uninstall

rm -vfr data certs
mkdir -p data certs indexes

./mkcert -install

cp $(mkcert -CAROOT)/rootCA.pem certs/ca.crt

htpasswd -cB -b auth.htpasswd user password

./mkcert -cert-file certs/registry.crt -key-file certs/registry.key localhost 127.0.0.1

sudo docker-compose up
