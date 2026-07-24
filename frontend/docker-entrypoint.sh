#!/bin/sh
if [ -n "$VM_DOMAIN" ] && [ -f /etc/nginx/nginx.prod.conf.template ]; then
  envsubst '${VM_DOMAIN}' < /etc/nginx/nginx.prod.conf.template > /etc/nginx/nginx.prod.conf
  exec nginx -c /etc/nginx/nginx.prod.conf -g "daemon off;"
else
  exec "$@"
fi
