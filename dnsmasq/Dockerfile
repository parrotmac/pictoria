FROM alpine:3.22.1

RUN apk add --no-cache dnsmasq bash iptables-legacy

RUN mkdir -p /etc/dnsmasq.d

COPY dnsmasq.conf /etc/dnsmasq.conf
COPY entrypoint.sh /entrypoint.sh
COPY bridge.sh /bridge.sh

# DHCP
EXPOSE 67/udp

ENTRYPOINT ["/entrypoint.sh"]
