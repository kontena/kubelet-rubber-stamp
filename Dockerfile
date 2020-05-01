FROM golang:1.11 as builder

RUN useradd -u 1001 kubelet-rubber-stamp

WORKDIR  /src

# Add dependency and download it
ADD go.mod .
ADD go.sum .
RUN go mod download

# Add source and compile
ADD . /src/

ARG ARCH=amd64
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} go build -a -installsuffix cgo -o kubelet-rubber-stamp cmd/manager/main.go


FROM scratch

COPY --from=builder /src/kubelet-rubber-stamp /kubelet-rubber-stamp
COPY --from=builder /etc/passwd /etc/passwd

USER 1001

ENTRYPOINT ["/kubelet-rubber-stamp", "-logtostderr"]
