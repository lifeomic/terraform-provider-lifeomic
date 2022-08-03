//go:build tools

package tools

// The below side-effect imports bring tools this project uses into it's go.mod
// so that they are versioned and don't need to be installed globally in $PATH.
// See: https://github.com/go-modules-by-example/index/blob/master/010_tools/README.md

import (
	// tfplugindocs generates and validates Terraform plugin docs.
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
)
