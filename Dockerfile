FROM golang:1.17 AS builder

# Turn on Go modules support and disable CGO
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Install upx (upx.github.io) to compress the compiled action
RUN apt-get -qq update && \
    apt-get -yqq install upx

WORKDIR src
COPY . .

# Compile the action
RUN go build \
  -a \
  -ldflags "-s -w -extldflags '-static'" \
  -installsuffix cgo \
  -tags netgo \
  -o /bin/action \
  .

# Strip any symbols - this is not a library
RUN strip /bin/action

# Compress the compiled action
RUN upx -q -9 /bin/action

RUN echo "nobody:x:65534:65534:Nobody:/:" > /etc_passwd




FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc_passwd /etc/passwd
COPY --from=builder --chown=65534:0 /bin/action /action

USER nobody
ENTRYPOINT ["/action"]
