FROM golang:latest
RUN curl https://glide.sh/get | sh
WORKDIR $GOPATH/src/github.com/ca-gip/artifactory-operator
COPY . $GOPATH/src/github.com/ca-gip/artifactory-operator
RUN make dep
RUN make build


FROM scratch
WORKDIR /root/
COPY --from=0 /go/src/github.com/ca-gip/artifactory-operator/build/artifactory-operator .
EXPOSE 8000
CMD ["./artifactory-operator"]
