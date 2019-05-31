# Makefile for squircy2

SUBPACKAGES := config data event eventsource irc plugin script sysinfo \
               web web/module web/module/* web/generated webhook

PLUGINS := $(patsubst plugins/%,%,$(wildcard plugins/*))

SOURCES := $(wildcard *.go) $(wildcard $(patsubst %,%/*.go,$(SUBPACKAGES)))
GENERATOR_SOURCES := $(wildcard web/views/*.twig) $(wildcard web/views/*/*.twig) $(wildcard web/public/css/*.css)

OUTPUT_BASE := out

PLUGIN_TARGETS := $(patsubst %,$(OUTPUT_BASE)/%.so,$(PLUGINS))
SQUIRCY_TARGET := $(OUTPUT_BASE)/squircy2

.PHONY: all build generate squircy2 plugins clean

all: build

clean:
	rm -rf $(OUTPUT_BASE)

build: generate plugins squircy2

generate: $(OUTPUT_BASE)/.generated

squircy2: $(SQUIRCY_TARGET)

plugins: $(PLUGIN_TARGETS)

run: build
	$(SQUIRCY_TARGET)

$(PLUGIN_TARGETS): $(OUTPUT_BASE)/%.so: $(SOURCES)
	go build -o $@ -buildmode=plugin plugins/$*/*.go

$(SQUIRCY_TARGET): $(SOURCES)
	go build -o $@ cmd/squircy2/*.go

$(OUTPUT_BASE)/.generated: $(GENERATOR_SOURCES)
	go generate
	touch $@

$(OUTPUT_BASE):
	mkdir -p $(OUTPUT_BASE)

$(SOURCES): $(OUTPUT_BASE)

$(GENERATOR_SOURCES): $(OUTPUT_BASE)
