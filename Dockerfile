### BUILDER ###
FROM golang:1.22.0-alpine3.19 AS builder

WORKDIR /src
COPY . .

# Build the binary
ENV GOOS=linux GOARCH=amd64
RUN go build -o /go/bin/server ./cmd/server

### FINAL IMAGE: server ###
FROM alpine:3.18 as server

COPY --from=builder /go/bin/server /go/bin/server

EXPOSE 8000/tcp

ENTRYPOINT ["/go/bin/server"]