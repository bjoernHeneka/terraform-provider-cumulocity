package main

import (
	"context"
	"flag"
	"log"

	"github.com/bjoernHeneka/terraform-provider-cumulocity/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// version is set during release builds via -ldflags.
var version string = "dev"

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "run provider with debugger support")
	flag.Parse()

	err := providerserver.Serve(context.Background(), provider.New(version), providerserver.ServeOpts{
		Address: "registry.terraform.io/bjoernHeneka/cumulocity",
		Debug:   debug,
	})
	if err != nil {
		log.Fatal(err)
	}
}
