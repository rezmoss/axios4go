# syntax=docker/dockerfile:1.19

# Scanner image for https://github.com/rezmoss/axios4go (OSS Scanner, https://github.com/anthropics/oss-scanner).
# The audit runs offline, so everything the build, tests and examples need is fetched and compiled here.

FROM golang:1.27-bookworm
ENV DEBIAN_FRONTEND=noninteractive
RUN sed -i 's|http://deb.debian.org|https://deb.debian.org|g' /etc/apt/sources.list.d/debian.sources \
    && apt-get update && apt-get install -y --no-install-recommends curl \
    && rm -rf /var/lib/apt/lists/*

# scanner contract: the checkout lives inside the image, at /src
COPY --exclude=.git --exclude=.oss-scanner/Dockerfile . /src
WORKDIR /src

RUN go mod download

# Library, tests (with and without -race) and every example.
RUN go build ./... \
    && go vet ./... \
    && go test -run '^$' ./... \
    && go test -race -run '^$' ./... \
    && mkdir -p .tmp/bin \
    && for d in examples/*/; do go build -o ".tmp/bin/$(basename "$d")" "./$d"; done

# Run the tests so the build log shows the image works, without failing the build.
RUN (timeout 900 go test -race ./... || echo "WARNING: go test reported failures")

# Everything is fetched. From here on the go command must not reach a proxy or download a toolchain.
ENV GOPROXY=off \
    GOTOOLCHAIN=local

CMD ["bash"]
