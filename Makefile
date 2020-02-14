
TEST_PKGS ?= ./...

.PHONY: all
all: test

.PHONY: test
test:
	go test -v $(TEST_PKGS)

.PHONY: bench
bench:
	go test -v -bench=. $(TEST_PKGS)
