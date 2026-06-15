# Build the manager binary
FROM --platform=$BUILDPLATFORM golang:1.26.4-alpine3.22@sha256:727cfc3c40be55cd1bc9a4a059406b28a059857e3be752aa9d09531e12c20c56 AS builder

ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

ARG GOPROXY=""
ENV GO111MODULE=on

WORKDIR /workspace

COPY go.mod go.mod
COPY go.sum go.sum

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download -x

# Copy the go source
COPY main.go main.go
COPY shutdown.go shutdown.go
COPY pkg/ pkg/

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -a -o eventsreader .

# Main container
FROM alpine:3.23.4@sha256:5b10f432ef3da1b8d4c7eb6c487f2f5a8f096bc91145e68878dd4a5019afde11

WORKDIR /events-reader/

COPY --from=builder /workspace/eventsreader /events-reader/

ENV USER_UID=1001 \
    USER_NAME=qubership-kube-events-reader
RUN adduser -u $USER_UID -DS $USER_NAME \
    && UID=$USER_NAME \
    && chown $UID /events-reader \
    && chmod -R 755 /events-reader

EXPOSE 8080

USER $USER_UID

CMD ["/events-reader/eventsreader"]
