FROM golang:1.16-alpine as golang
RUN apk add -U --no-cache ca-certificates git make
COPY . /src
WORKDIR /src
RUN FLAVOR=stable CGO_ENABLED=0 GOPROXY=direct make

FROM alpine
COPY --from=golang /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=golang /src/ebs-optimizer .
ENTRYPOINT ["./ebs-optimizer"]
