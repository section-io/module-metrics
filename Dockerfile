FROM golang:1.19

ENV CGO_ENABLED=0

RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.48.0

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN golangci-lint run ./...
RUN go test -short -v ./...
