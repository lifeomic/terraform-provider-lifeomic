PROJECT := terraform-provider-phc
PKG     := github.com/lifeomic/$(PROJECT)

GIT_HEAD := $(shell git rev-list --all | head -n 1)
GIT_REF  ?= dev

LD_FLAGS ?= "-X $(REPO)/internal/client.GitCommit=$(GIT_HEAD) -X $(REPO)/internal/client.GitRef=$(GIT_REF)"

ACC_TEST_COUNT       ?= 1
ACC_TEST_PARALLELISM ?= 10
ACC_TEST_PKG         := ./internal/provider/...

clean:
	rm $(PROJECT)

build: clean
	go build -ldflags=$(LD_FLAGS) -o $(PROJECT) main.go

$(PROJECT): build

unittest:
	go test -v $(TESTARGS) ./...

acctest:
	TF_ACC=1 go test -v $(ACC_TEST_PKG) $(TESTARGS) -count $(ACC_TEST_COUNT) -parallel $(ACC_TEST_PARALLELISM) -ldflags=$(LD_FLAGS)

test: unittest acctest

generate:
	go generate ./...

.PHONY: build clean unittest acctest test generate
