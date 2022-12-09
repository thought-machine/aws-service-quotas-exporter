# golang:1.19.4-alpine3.17
FROM golang@sha256:f33331e12ca70192c0dbab2d0a74a52e1dd344221507d88aaea605b0219a212f as builder

RUN ln -s /usr/local/go/bin/go /usr/local/bin/go

RUN apk add --no-cache curl wget gcc make bash git musl-dev libc6-compat

WORKDIR /go/github.com/thought-machine/aws-service-quotas-exporter

COPY . .

RUN go build -ldflags "-linkmode external -extldflags -static" -o aws-service-quotas-exporter cmd/main.go

# Max user
RUN addgroup -g 255999 -S app && \
    adduser -u 255999 -S app -G app

USER app

FROM scratch

EXPOSE 9090

COPY --from=builder /go/github.com/thought-machine/aws-service-quotas-exporter/aws-service-quotas-exporter /home/app/aws-service-quotas-exporter

COPY --from=0 /etc/passwd /etc/passwd
USER app

ENTRYPOINT ["/home/app/aws-service-quotas-exporter"]
