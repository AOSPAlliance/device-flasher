PROGRAM_NAME ?= device-flasher
EXTENSIONS := linux exe darwin
NAMES := $(PROGRAM_NAME)
PROGRAMS := $(foreach PROG,$(NAMES),$(foreach EXT,$(EXTENSIONS),$(PROG).$(EXT)))
VERSION := $(shell git describe --always --tags --dirty='-dirty')
LDFLAGS := -ldflags "-X main.version=$(VERSION)"
COMMON_ARGS := GOARCH=amd64 CGO_ENABLED=0

all: clean build

$(PROGRAM_NAME).linux:
	$(COMMON_ARGS) GOOS=linux go build -mod=vendor $(TAGS) $(LDFLAGS) -o $@

$(PROGRAM_NAME).exe:
	$(COMMON_ARGS) GOOS=windows go build -mod=vendor $(TAGS) $(LDFLAGS) -o $@

$(PROGRAM_NAME).darwin:
	$(COMMON_ARGS) GOOS=darwin go build -mod=vendor $(TAGS) $(LDFLAGS) -o $@

.PHONY: build
build: $(PROGRAMS)
	@echo Built $(VERSION)

clean:
	-rm $(PROGRAMS)

test:
	go test -v ./...
