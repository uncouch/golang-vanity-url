FROM golang:1.15 as builder

COPY . /workspace
WORKDIR /workspace

# build a static binary for service
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -a -o /workspace/service.bin main.go

# Build distribution image
FROM gcr.io/distroless/base:nonroot
COPY --from=builder /workspace/service.bin /usr/local/bin/service
USER nonroot
CMD ["/usr/local/bin/service"]
