FROM --platform=linux/amd64 debian:stable-slim

RUN apt-get update && apt-get install -y ca-certificates

ADD adventrak /usr/bin/adventrak
WORKDIR /app
COPY .env /app/.env

CMD ["adventrak"]