package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/lifeomic/terraform-provider-phc/internal/provider"
)

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "Set true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/lifeomic/phc",
		Debug:   debugMode,
	}

	err := providerserver.Serve(context.Background(), provider.New, opts)

	if err != nil {
		log.Fatalf("failed to initialize provider: %v\n", err)
	}
}
