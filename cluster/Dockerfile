# Build the manager binary
FROM golang:1.26.2@sha256:5f3787b7f902c07c7ec4f3aa91a301a3eda8133aa32661a3b3a3a86ab3a68a36 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd/ cmd/
COPY apis/ apis/
COPY internal/ internal/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o provider cmd/provider/main.go

FROM alpine:3.23@sha256:5b10f432ef3da1b8d4c7eb6c487f2f5a8f096bc91145e68878dd4a5019afde11
WORKDIR /
COPY --from=builder /workspace/provider .
COPY package /

EXPOSE 8080
ENTRYPOINT ["/provider"]
