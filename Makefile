# used to generate grpc code.
PROTO_DIR := proto
PROTO_SRC := $(shell find $(PROTO_DIR) -name "*.proto")
GO_OUT := .

.PHONY: generate-proto
generate-proto:
	protoc \
		--proto_path=$(PROTO_DIR) \
		--go_out=$(GO_OUT) \
		--go-grpc_out=$(GO_OUT) \
		$(PROTO_SRC)
