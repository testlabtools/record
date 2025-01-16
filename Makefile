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

.PHONY: deps go-deps

deps: go-deps

go-deps:
	go mod download
	cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -P5 -tI % go install %
