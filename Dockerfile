FROM golang:1.14-alpine as builder

ADD . /workspace
WORKDIR /workspace

ENV GOPROXY=https://proxy.golang.org
ENV GOPATH="/go"

# 'go install' puts the rotation binary in /go/bin/
RUN go env && go install -ldflags="-s -w" rotation.go

# -------

FROM alpine:latest
RUN apk add --no-cache ca-certificates

COPY --from=builder /go/bin/rotation /usr/local/bin/rotation

ENTRYPOINT ["rotation"]
