package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("NOVU_API_KEY"); v == "" {
		t.Fatal("NOVU_API_KEY must be set for acceptance tests")
	}
}

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"novu": providerserver.NewProtocol6WithError(New("test")()),
}
