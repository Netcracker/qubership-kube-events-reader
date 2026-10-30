# Build the manager binary
FROM --platform=$BUILDPLATFORM golang:1.27.2-alpine3.24@sha256:f92b6ef800e499660581efdabdf25d9d817a9d124eaf900924f0504e7e27e12d AS builder

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
FROM alpine:3.24.3@sha256:be2656061b536c1883c4034790e21d0a09c5aaaf51968b89368a3c0371ea2c9a

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
