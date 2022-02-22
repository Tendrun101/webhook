FROM golang:1.16 AS builder

WORKDIR /go/src/pod-mutate-webhook

COPY ../* ./
RUN CGO_ENABLED=0 go build -o /go/bin/pod-mutate-webhook

FROM alpine:3.14
COPY --from=builder /go/bin/pod-mutate-webhook /pod-mutate-webhook
ENTRYPOINT ["/pod-mutate-webhook"]