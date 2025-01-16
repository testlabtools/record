# Disable CGO for all builds.
export CGO_ENABLED=0

.PHONY: build test generate clean

build: | generate
	go build ${GOFLAGS} -o dist/main

test: | generate
	go test ./...

generate:
	go generate ./...

clean:
	rm -fvr dist
