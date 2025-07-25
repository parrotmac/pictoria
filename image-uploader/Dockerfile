FROM alpine:3.22.1

RUN apk add --no-cache s3cmd bash

COPY s3cmd-config /root/.s3cfg
COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT [ "/entrypoint.sh" ]
