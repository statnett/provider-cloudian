# Build the manager binary
FROM golang:1.25@sha256:c0bf2bc2f8e5720aa2e83e45d2085edbf2ad085e2d1a195bb6c3c402350fe661 as builder

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

FROM alpine:3.22@sha256:4b7ce07002c69e8f3d704a9c5d6fd3053be500b7f1c69fc0d80990c2ad8dd412
WORKDIR /
COPY --from=builder /workspace/provider .
COPY package /

EXPOSE 8080
ENTRYPOINT ["/provider"]
