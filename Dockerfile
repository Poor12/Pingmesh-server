FROM golang:1.14.2 as build
ENV CGO_ENABLED=0
ENV GOPATH=/go
ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /go/src/sigs.k8s.io/pingmesh-server
COPY go.mod .
COPY go.sum .
COPY vendor vendor
RUN go mod download

COPY pkg pkg
COPY cmd cmd

ARG GOARCH
ARG LDFLAGS
RUN go build -mod=readonly -ldflags "$LDFLAGS" -o /pingmesh-server $PWD/cmd/pingmesh-server

FROM golang:1.14.2

COPY --from=build pingmesh-server /

USER 65534

ENTRYPOINT ["/pingmesh-server"]