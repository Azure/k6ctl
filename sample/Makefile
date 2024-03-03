BIN_DIR := $(shell pwd)/../bin

build: build-plugin-hello

build-plugin-hello:
	$(call run-in-folder,plugin-hello,go build -o $(BIN_DIR)/k6ctl-hello)

define run-in-folder
[ -d $(1) ] && { \
	set -e ;\
	echo "Running $(2) in $(1)"; \
	cd $(1); \
	$(2); \
	cd ..; \
}
endef