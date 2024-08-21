FROM golang:1.23

ENV CGO_ENABLED=0

RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.60.1

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -v ./...
RUN golangci-lint run ./...
RUN go test -short -v ./...
