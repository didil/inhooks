MYGOBIN = $(PWD)/bin
GIT_SHA = $(shell git rev-parse --short HEAD)

install-tools:
	@echo MYGOBIN: $(MYGOBIN)
	@cat tools/tools.go | grep _ | awk -F'"' '{print $$2}' | GOBIN=$(MYGOBIN) xargs -tI % go install %

install:
	@echo MYGOBIN: $(MYGOBIN)
    GOBIN=$(MYGOBIN) go install ./...

test:
	go test -race ./...

test-coverage:
	go test -race -covermode atomic -coverprofile=coverprofile.txt ./...

lint:
	$(MYGOBIN)/golangci-lint run

build:
	go build -ldflags="-X 'main.version=$(GIT_SHA)'" -o bin/api cmd/api/main.go

run-dev:
	APP_ENV=development go run cmd/api/main.go

.PHONY: gen-mocks
gen-mocks:
	pkg/testsupport/mocks/gen_mocks.sh

goreleaser-snapshot:
	goreleaser --snapshot --clean
