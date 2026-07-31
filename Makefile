BINARY := aint
CMD := ./cmd/aint

.PHONY: build test vet fmt install clean

build:
	go build -o $(BINARY) $(CMD)

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -l .

install:
	go install $(CMD)

clean:
	rm -f $(BINARY)
