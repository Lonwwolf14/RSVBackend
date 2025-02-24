FROM --platform=linux/amd64 debian:stable-slim

RUN apt-get update && apt-get install -y ca-certificates

COPY rsvbackend /usr/bin/rsvbackend
COPY templates /templates

RUN chmod +x /usr/bin/rsvbackend

CMD ["rsvbackend"]