FROM caddy:2-builder-alpine AS builder

RUN xcaddy build \
    --with github.com/tailscale/caddy-tailscale

FROM caddy:2-alpine

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
COPY Caddyfile /etc/caddy/Caddyfile

# sh -c "echo 'nameserver 1.1.1.1' >> /etc/resolv.conf && caddy reverse-proxy --to https://pictoria.fluffy-broadnose.ts.net --from http://:8080 --change-host-header --debug --access-log"