FROM golang:1.17

ENV CGO_ENABLED=0

# install tooling
RUN go get -v \
  github.com/kisielk/errcheck \
  golang.org/x/lint/golint

# Using a path outside the GOPATH is the convention to trigger Go Modules semantics.
#  https://github.com/golang/go/wiki/Modules#how-to-use-modules
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN go test -short -v ./...

RUN gofmt -e -s -d . 2>&1 | tee /gofmt.out && test ! -s /gofmt.out
RUN go vet .
RUN golint -set_exit_status
RUN errcheck ./...
