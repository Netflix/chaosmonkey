BINARY=chaosmonkey
GOARCH=amd64

COMMIT=$(shell git rev-parse HEAD)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
REPO=$(shell basename $(shell git rev-parse --show-toplevel))
PKGS := $(shell go list ./... | grep -v -e /integration -e /vendor)
INTEGRATION_PKGS := $(shell go list ./... | grep /integration)

BUILD_DIR=$(shell pwd)/build
CURRENT_DIR=$(shell pwd)

LDFLAGS = -ldflags "-X main.COMMIT=${COMMIT} -X main.BRANCH=${BRANCH}"
GOFLAGS='-mod=vendor' 
GOOS=linux 
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
        LDFLAGS = -ldflags "-X main.COMMIT=${COMMIT} -X main.BRANCH=${BRANCH} -linkmode external -extldflags -static -s -w"
endif

all: clean test build

docker:
	docker build . -t netflix/chaosmonkey:${COMMIT}

clean:
	rm -rf ${BUILD_DIR}
	go clean

build: check
	go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY} ./cmd/${BINARY}/main.go

check: fmt  

gofmt: fmt

fmt: 
	go fmt $$(go list ./... | grep -v /vendor/) ; 

lint: $(GOLINT)
	golangci-lint run 

#errcheck:
	#errcheck -ignore 'io:Close' -ignoretests `go list ./... | grep -v -e '/vendor/' -e '/migration'`

test:
	go test -v ./...


# Coverage testing
cover:
	echo 'mode: atomic' > coverage.out 
	go list ./... | grep -Ev '/vendor/|/migration' | xargs -n1 -I{} sh -c 'go test -covermode=atomic -coverprofile=coverage.tmp {} && tail -n +2 coverage.tmp >> coverage.out' && rm coverage.tmp
	go tool cover -html=coverage.out

fix:
	gofmt -w -s `find . -name '*.go' | grep -Ev '/vendor/|/migration'`
