# Run:
# docker build -t golang-builder -f Dockerfile.builder .
# to create golang-builder
docker run --rm -v "$PWD":/go/src/github.com/tsuru/tsuru -w /go/src/github.com/tsuru/tsuru goland-builder:latest make binaries 
