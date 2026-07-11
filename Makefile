# used to generate grpc code.
# to run the code: make generate-proto DIR=proto/sub-dir
PROTO_DIR := proto
GO_OUT := .

.PHONY: generate-proto

generate-proto:
	@test -n "$(DIR)" || (echo "Usage: make generate-proto DIR=proto/user"; exit 1)

	protoc \
		-I=$(PROTO_DIR) \
		--go_out=$(GO_OUT) \
		--go-grpc_out=$(GO_OUT) \
		$(shell find $(DIR) -type f -name "*.proto")
