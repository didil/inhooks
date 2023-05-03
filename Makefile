MYGOBIN = $(PWD)/bin

install-tools:
	@echo MYGOBIN: $(MYGOBIN)
	@cat tools/tools.go | grep _ | awk -F'"' '{print $$2}' | GOBIN=$(MYGOBIN) xargs -tI % go install %

install:
	@echo MYGOBIN: $(MYGOBIN)
    GOBIN=$(MYGOBIN) go install ./...
test:
	go test -race ./...

lint:
	$(MYGOBIN)/golangci-lint run

build:
	go build -o bin/api cmd/api/main.go

run-dev:
	APP_ENV=development go run cmd/api/main.go

.PHONY: gen-mocks
gen-mocks:
	pkg/testsupport/mocks/gen_mocks.sh