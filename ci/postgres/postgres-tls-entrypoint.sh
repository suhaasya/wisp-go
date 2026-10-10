#!/bin/bash
# Postgres rejects ssl_key_file when permissions are too open (git checkout is often 0644).
set -euo pipefail
dest=/var/lib/postgresql/wisp-certs
mkdir -p "$dest"
cp /certs/server.pem /certs/ca.pem "$dest/"
cp /certs/server.key "$dest/"
chmod 600 "$dest/server.key"
chown -R postgres:postgres "$dest"
exec docker-entrypoint.sh postgres \
  -c ssl=on \
  -c "ssl_cert_file=${dest}/server.pem" \
  -c "ssl_key_file=${dest}/server.key" \
  -c "ssl_ca_file=${dest}/ca.pem"
