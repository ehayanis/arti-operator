FROM golang:latest as build
RUN curl https://glide.sh/get | sh && glide --version
WORKDIR $GOPATH/src/github.com/ca-gip/artifactory-operator
COPY . $GOPATH/src/github.com/ca-gip/artifactory-operator
RUN make dep
RUN make build

FROM alpine:latest as certificates-source
RUN apk update && apk add ca-certificates
COPY assets/certificates /usr/local/share/ca-certificates/
RUN update-ca-certificates

FROM alpine
WORKDIR /root/
COPY --from=build /go/src/github.com/ca-gip/artifactory-operator/build/artifactory-operator .
COPY --from=certificates-source /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 8000
CMD ["./artifactory-operator"]
