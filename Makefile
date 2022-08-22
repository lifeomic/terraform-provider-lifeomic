PROJECT := terraform-provider-lifeomic
PKG     := github.com/lifeomic/$(PROJECT)

GQL_PKG              := ./internal/gqlclient
GQL_SCHEMA_DIR       := $(GQL_PKG)/schemas
GQL_GENERATED_SCHEMA := $(GQL_PKG)/schema.graphql

GIT_HEAD := $(shell git rev-list --all | head -n 1)
GIT_REF  ?= dev

LD_FLAGS ?= "-X $(REPO)/internal/client.GitCommit=$(GIT_HEAD) -X $(REPO)/internal/client.GitRef=$(GIT_REF)"

ACC_TEST_COUNT       ?= 1
ACC_TEST_PARALLELISM ?= 10
ACC_TEST_PKG         := ./internal/provider/...

default: clean generate build

$(GQL_SCHEMA_DIR):
	mkdir -p $(GQL_SCHEMA_DIR)
	cd $(GQL_PKG) && lifeomic-fetch-individual-graphql-schemas -s marketplace,marketplaceAuthed,appStore

$(GQL_GENERATED_SCHEMA):
	cd $(GQL_PKG) && yarn && yarn graphql-codegen

$(GQL_PKG)/generated.go: generate

$(PROJECT): build

clean:
	rm -rf $(PROJECT) $(GQL_SCHEMA_DIR) $(GQL_GENERATED_SCHEMA)

build:
	go build -ldflags=$(LD_FLAGS) -o $(PROJECT) main.go

unittest:
	go test -v $(TESTARGS) ./...

acctest:
	TF_ACC=1 go test -v $(ACC_TEST_PKG) $(TESTARGS) -count $(ACC_TEST_COUNT) -parallel $(ACC_TEST_PARALLELISM) -ldflags=$(LD_FLAGS)

test: unittest acctest

generate-docs:
	go generate main.go

generate: $(GQL_SCHEMA_DIR) $(GQL_GENERATED_SCHEMA)
	go generate ./...

.PHONY: build clean unittest acctest test generate generate-docs

