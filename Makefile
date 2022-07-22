PROJECT := terraform-provider-phc
REPO    := github.com/lifeomic/$(PROJECT)

GIT_HEAD := $(shell git rev-list --all | head -n 1)
GIT_REF  ?= dev

LD_FLAGS ?= "-X $(REPO)/internal/client.GitCommit=$(GIT_HEAD) -X $(REPO)/internal/client.GitRef=$(GIT_REF)"

build:
	go build -ldflags=$(LD_FLAGS) -o $(PROJECT) main.go

