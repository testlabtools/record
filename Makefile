GO_TEST_FLAGS ?= -test.timeout 2s

# Disable CGO for all builds.
export CGO_ENABLED=0

.PHONY: build test cov generate clean

build: | generate
	go build ${GOFLAGS} -o dist/main ./cmd

test: | generate
	go test ${GO_TEST_FLAGS} ./...

cov: GO_TEST_FLAGS+=-coverprofile=coverage.out
cov: test
	go tool cover -html=coverage.out

generate:
	go generate ./...

clean:
	rm -fvr dist testdata/*/repo

.PHONY: deps go-deps fake-repos

deps: go-deps fake-repos

go-deps:
	go mod download
	cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -P5 -tI % go install %

fake-repos: testdata/github/repo/.git

%/.git:
	./scripts/fake-repo.sh $(@D)
