# to build this docker image:
#   docker build .
FROM ghcr.io/hybridgroup/opencv:4.11.0

ENV GOPATH /go

COPY . /go/src/gocv.io/x/gocv/

WORKDIR /go/src/gocv.io/x/gocv
RUN go build -o /build/gocv_version ./main.go

CMD ["/build/gocv_version"]