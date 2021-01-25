package main

import (
	"terraform-provider-powerdnsadmin/powerdnsadmin"

	"github.com/hashicorp/terraform-plugin-sdk/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: powerdnsadmin.Provider})
}
