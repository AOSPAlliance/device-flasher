PROGRAM_NAME ?= device-flasher
PROGRAM_GUI_NAME := $(PROGRAM_NAME)-gui
EXTENSIONS := linux exe darwin
NAMES := $(PROGRAM_NAME) $(PROGRAM_GUI_NAME)
PROGRAMS := $(foreach PROG,$(NAMES),$(foreach EXT,$(EXTENSIONS),$(PROG).$(EXT)))
VERSION := $(shell git describe --always --tags --dirty='-dirty')
LDFLAGS := -ldflags "-X main.version=$(VERSION)"
COMMON_ARGS := GOARCH=amd64

$(PROGRAM_NAME).%: CGO := CGO_ENABLED=0
$(PROGRAM_NAME).%: TAGS := -tags ""
$(PROGRAM_GUI_NAME).%: CGO := CGO_ENABLED=1
$(PROGRAM_GUI_NAME).%: TAGS := -tags GUI

all: clean build

# CLI, default
$(PROGRAM_NAME).linux:
	$(COMMON_ARGS) $(CGO) GOOS=linux go build -mod=vendor $(TAGS) $(LDFLAGS) -o $@

$(PROGRAM_NAME).exe:
	$(COMMON_ARGS) $(CGO) GOOS=windows go build -mod=vendor $(TAGS) $(LDFLAGS) -o $@

$(PROGRAM_NAME).darwin:
	$(COMMON_ARGS) $(CGO) GOOS=darwin go build -mod=vendor $(TAGS) $(LDFLAGS) -o $@

# GUI
$(PROGRAM_GUI_NAME).linux:
	$(COMMON_ARGS) $(CGO) GOOS=linux go build -mod=vendor $(TAGS) $(LDFLAGS) -o $@

$(PROGRAM_GUI_NAME).exe:
	$(COMMON_ARGS) $(CGO) GOOS=windows go build -mod=vendor $(TAGS) $(LDFLAGS) -o $@

$(PROGRAM_GUI_NAME).darwin:
	$(COMMON_ARGS) $(CGO) GOOS=darwin go build -mod=vendor $(TAGS) $(LDFLAGS) -o $@

.PHONY: build
build: $(PROGRAMS)
	@echo Built $(VERSION)

clean:
	-rm $(PROGRAMS)

test:
	go test -v ./...
