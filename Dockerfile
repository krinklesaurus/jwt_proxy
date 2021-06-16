############################
# STEP 1 build executable binary
############################
# golang alpine 1.14.4
FROM golang:1.16.5-alpine3.12 as builder

# Install git + SSL ca certificates.
# Git is required for fetching the dependencies.
# Ca-certificates is required to call HTTPS endpoints.
RUN apk update && apk add --no-cache git ca-certificates tzdata build-base && update-ca-certificates

# Create appuser
ENV USER=appuser
ENV UID=10001

# See https://stackoverflow.com/a/55757473/12429735
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

WORKDIR /jwt-proxy
COPY . .

# Fetch dependencies.
RUN go mod download &&\
    go mod verify

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o jwt-proxy .


############################
# STEP 2 build a small image
############################
FROM alpine:3.14

RUN apk upgrade --update-cache

# Import from builder.
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

WORKDIR /jwt-proxy

COPY --from=builder /jwt-proxy/jwt-proxy .
COPY --from=builder /jwt-proxy/certs certs
COPY --from=builder /jwt-proxy/www www
COPY --from=builder /jwt-proxy/config.yml config.yml

# Use an unprivileged user.
USER appuser:appuser

ENTRYPOINT ["./jwt-proxy"]

EXPOSE 8080