FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ .
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /server /server
EXPOSE 8080
ENTRYPOINT ["/server"]
