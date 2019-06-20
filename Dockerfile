FROM golang:1.12

ENV CGO_ENABLED=0

# install tooling
RUN go get -v \
  github.com/kisielk/errcheck \
  golang.org/x/lint/golint

# Using a path outside the GOPATH is the convention to trigger Go Modules semantics.
#  https://github.com/golang/go/wiki/Modules#how-to-use-modules
WORKDIR /src

# BEGIN pre-install dependencies to reduce time for each code change to build
COPY go.mod go.sum ./
# Using `go get` currently seems to miss some nested dependencies.
RUN go mod download
# END

COPY . .

RUN gofmt -e -s -d . 2>&1 | tee /gofmt.out && test ! -s /gofmt.out
RUN go vet .
RUN golint -set_exit_status

RUN errcheck ./...

RUN go install ./...

RUN go test -v ./...