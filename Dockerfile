FROM golang:1.11 as builder

RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

WORKDIR  /go/src/github.com/kontena/kubelet-rubber-stamp

# Add dependency graph and vendor it in
ADD Gopkg.* /go/src/github.com/kontena/kubelet-rubber-stamp/
RUN dep ensure -v -vendor-only

# Add source and compile
ADD . /go/src/github.com/kontena/kubelet-rubber-stamp/

ARG ARCH=amd64
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} go build -a -installsuffix cgo -o kubelet-rubber-stamp cmd/manager/main.go


FROM scratch

COPY --from=builder /go/src/github.com/kontena/kubelet-rubber-stamp/kubelet-rubber-stamp /kubelet-rubber-stamp

ENTRYPOINT ["/kubelet-rubber-stamp", "-logtostderr"]
