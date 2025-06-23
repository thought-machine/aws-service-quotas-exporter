FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS builder

ARG TARGETARCH

RUN ln -s /usr/local/go/bin/go /usr/local/bin/go

RUN apk add --no-cache curl wget gcc make bash git musl-dev libc6-compat

WORKDIR /go/github.com/thought-machine/aws-service-quotas-exporter

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -o aws-service-quotas-exporter cmd/main.go

# Max user
RUN addgroup -g 255999 -S app && \
    adduser -u 255999 -S app -G app

USER app

FROM scratch

EXPOSE 9090

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/github.com/thought-machine/aws-service-quotas-exporter/aws-service-quotas-exporter /home/app/aws-service-quotas-exporter

COPY --from=0 /etc/passwd /etc/passwd
USER app

ENTRYPOINT ["/home/app/aws-service-quotas-exporter"]
