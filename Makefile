################################################################################
# This Makefile contains useful commands to develop with

# Run all tests with the race flag
test:
	go test -race ./pkg/...

# Views test coverage as a pretty HTML document
coverage: coverage.out
	@go tool cover -func=coverage.out
	@go tool cover -html=coverage.out

# Run go fmt on all our stuff
fmt:
	@go fmt ./pkg/...

################################################################################
# Dependencies
GO_FILES = $(shell find . -type f -name '*.go')

coverage.out: $(GO_FILES)
	go test -coverprofile=coverage.out ./...

