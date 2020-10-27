################################################################################
# This Makefile contains useful commands to develop with

test:
	go test -race ./pkg/...

fmt:
	@go fmt ./pkg/...

