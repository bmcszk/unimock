FROM golang:1.24-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o /unimock .

FROM gcr.io/distroless/static-debian12
COPY --from=builder /unimock /usr/local/bin/unimock
COPY config.yaml /etc/unimock/config.yaml
ENV UNIMOCK_PORT=8080 \
    UNIMOCK_CONFIG=/etc/unimock/config.yaml \
    UNIMOCK_LOG_LEVEL=info
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/unimock"] 
