FROM golang:1.23

ENV CGO_ENABLED=0

RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.60.1

RUN go env GOCACHE | grep -xF '/root/.cache/go-build'
RUN go env GOMODCACHE | grep -xF '/go/pkg/mod'

WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod/ \
  --mount=type=cache,target=/root/.cache/go-build/ \
  go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod/ \
  --mount=type=cache,target=/root/.cache/go-build/ \
  go build -v ./...

RUN --mount=type=cache,target=/go/pkg/mod/ \
  --mount=type=cache,target=/root/.cache/go-build/ \
  golangci-lint run ./...

RUN --mount=type=cache,target=/go/pkg/mod/ \
  --mount=type=cache,target=/root/.cache/go-build/ \
  go test -short -v ./...

RUN --mount=type=cache,target=/go/pkg/mod/ \
  --mount=type=cache,target=/root/.cache/go-build/ \
  go test -run=BenchmarkOnlyNoTests -bench=. ./...
