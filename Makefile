BINARY := aint
CMD := ./cmd/aint

.PHONY: build test test-race vet fmt fmt-check aint-check lint install clean

build:
	go build -o $(BINARY) $(CMD)

test:
	go test ./...

test-race:
	go test -race ./...

vet:
	go vet ./...

fmt:
	gofmt -l .

fmt-check:
	@test -z "$$(gofmt -l .)" || (gofmt -l . && exit 1)

aint-check:
	go run $(CMD) check .

lint: fmt-check vet test-race aint-check

install:
	go install $(CMD)

clean:
	rm -f $(BINARY)
