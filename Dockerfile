FROM golang:1.17 as builder
RUN useradd -u 10001 app
ADD . /build/
WORKDIR /build/cmd/speedtest/
RUN CGO_ENABLED=0 GOOS=linux go build -a -o speedtest .

FROM alpine
# required to have non-privileged user
COPY --from=builder /etc/passwd /etc/passwd
# required to access Internet
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /app
COPY cmd/speedtest/reference_env_walkaround.conf reference.conf
COPY --from=builder /build/cmd/speedtest/speedtest speedtest

USER app
CMD [ "/app/speedtest" ]