VERSION := $(shell git rev-parse --short HEAD)
BUILDFLAGS := -ldflags "-X main.version=$(VERSION)"

build:
	go build -v $(BUILDFLAGS)
