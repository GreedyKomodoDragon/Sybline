# Build stage
FROM golang:alpine AS builder

RUN apk add --no-cache git && mkdir /app
ADD . /app
WORKDIR /app

# Fetch dependencies
RUN go mod tidy && go build --ldflags '-w -s' -o main cmd/main.go

# Final stage
FROM registry.access.redhat.com/ubi8/ubi-minimal:8.8-1072

# Copy the built binary from the builder stage
COPY --from=builder /app/main /app/main

USER 1001

CMD ["/app/main"]
