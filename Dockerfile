# Build the manager binary
FROM --platform=$BUILDPLATFORM golang:1.24.2-alpine3.21 AS builder

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
FROM alpine:3.22.0

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
