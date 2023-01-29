FROM golang:1.18-alpine AS builder

WORKDIR /src

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
RUN mv tagbot /bin/.

FROM alpine:3.17
COPY --from=builder /bin/tagbot /bin/tagbot
ENTRYPOINT ["/bin/tagbot"]
