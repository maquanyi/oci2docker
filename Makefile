BUILDTAGS=
export GOPATH:=$(CURDIR)/Godeps/_workspace:$(GOPATH)

all:
	go build -tags "$(BUILDTAGS)" -o oci2docker .

install:
	cp oci2docker /usr/local/bin/oci2docker

clean:
	go clean

.PHONY: test .gofmt .govet .golint

test: .gofmt .govet .golint

.gofmt:
	go fmt ./...

.govet:
	go vet -x ./...

.golint:
	golint ./...
