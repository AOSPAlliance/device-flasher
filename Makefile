PROGRAM_NAME ?= device-flasher
PROGRAM_GUI_NAME := $(PROGRAM_NAME)-gui
PROGRAM_GUI_CALYXOS_NAME := $(PROGRAM_NAME)-CalyxOS-gui
EXTENSIONS := linux exe darwin
NAMES := $(PROGRAM_NAME) $(PROGRAM_GUI_NAME) $(PROGRAM_GUI_CALYXOS_NAME)
PROGRAMS := $(foreach PROG,$(NAMES),$(foreach EXT,$(EXTENSIONS),$(PROG).$(EXT)))
VERSION := $(shell git describe --always --tags --dirty='-dirty')
LDFLAGS := -ldflags "-X main.version=$(VERSION)"
COMMON_ARGS := GOARCH=amd64
SOURCES := $(wildcard *.go internal/*/*.go resources/*.go resources/*/*.go)
RESOURCES := $(patsubst %.png,%.go,$(wildcard resources/*.png))
RESOURCES += $(patsubst %.svg,%.go,$(wildcard resources/*.svg))
CALYXOS_RESOURCES := $(patsubst %.png,%.go,$(wildcard resources/calyxos/*.png))
CALYXOS_RESOURCES += $(patsubst %.svg,%.go,$(wildcard resources/calyxos/*.svg))

$(PROGRAM_NAME).%: CGO := CGO_ENABLED=0
$(PROGRAM_NAME).%: TAGS := -tags ""
$(PROGRAM_GUI_NAME).%: CGO := CGO_ENABLED=1
$(PROGRAM_GUI_NAME).%: TAGS := -tags "GUI en"
$(PROGRAM_GUI_CALYXOS_NAME).%: CGO := CGO_ENABLED=1
$(PROGRAM_GUI_CALYXOS_NAME).%: TAGS := -tags "GUI en CalyxOS"

all: build

# CLI, default
$(PROGRAM_NAME).linux: $(SOURCES)
	$(COMMON_ARGS) $(CGO) GOOS=linux go build -mod=vendor $(TAGS) $(LDFLAGS) -o $@

$(PROGRAM_NAME).exe: $(SOURCES)
	$(COMMON_ARGS) $(CGO) GOOS=windows go build -mod=vendor $(TAGS) $(LDFLAGS) -o $@

$(PROGRAM_NAME).darwin: $(SOURCES)
	$(COMMON_ARGS) $(CGO) GOOS=darwin go build -mod=vendor $(TAGS) $(LDFLAGS) -o $@

# GUI
$(PROGRAM_GUI_NAME).linux: $(SOURCES)
	fyne-cross linux -arch amd64 -env "GOFLAGS=-mod=vendor" $(TAGS) $(LDFLAGS) -name $@

$(PROGRAM_GUI_NAME).exe: $(SOURCES)
	fyne-cross windows -arch amd64 -env "GOFLAGS=-mod=vendor" $(TAGS) $(LDFLAGS) -name $@

$(PROGRAM_GUI_NAME).darwin: $(SOURCES)
	fyne-cross darwin -arch amd64 -env "GOFLAGS=-mod=vendor" $(TAGS) $(LDFLAGS) -name $@

# CalyxOS GUI
$(PROGRAM_GUI_CALYXOS_NAME).linux: $(SOURCES)
	fyne-cross linux -arch amd64 -env "GOFLAGS=-mod=vendor" $(TAGS) $(LDFLAGS) -name $@

$(PROGRAM_GUI_CALYXOS_NAME).exe: $(SOURCES)
	fyne-cross windows -arch amd64 -env "GOFLAGS=-mod=vendor" $(TAGS) $(LDFLAGS) -name $@

$(PROGRAM_GUI_CALYXOS_NAME).darwin: $(SOURCES)
	fyne-cross darwin -arch amd64 -env "GOFLAGS=-mod=vendor" $(TAGS) $(LDFLAGS) -name $@

resources/%.go: resources/%.png
	fyne bundle -package resources $< > $@
# Hack since we have the resources in a different package
	sed -i 's/var resource/var Resource/' $@

resources/%.go: resources/%.svg
	fyne bundle -package resources $< > $@
# Hack since we have the resources in a different package
	sed -i 's/var resource/var Resource/' $@

resources/calyxos/%.go: resources/calyxos/%.png
	fyne bundle -package calyxos $< > $@
# Hack since we have the resources in a different package
	sed -i 's/var resource/var Resource/' $@

resources/calyxos/%.go: resources/calyxos/%.svg
	fyne bundle -package calyxos $< > $@
# Hack since we have the resources in a different package
	sed -i 's/var resource/var Resource/' $@

.PHONY: build
build: $(PROGRAMS)
	@echo Built $(VERSION)

.PHONY: resources
resources: $(RESOURCES) $(CALYXOS_RESOURCES)
	@echo Updated resources

clean:
	-rm $(PROGRAMS)

test:
	go test -v ./...
