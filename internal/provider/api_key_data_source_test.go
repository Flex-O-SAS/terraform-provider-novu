package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestApiKeyDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: `
				data "novu_api_key" "test" {
				}
				`,
				ExpectError: regexp.MustCompile("Multiple api keys found"),
			},
			{
				Config: testApiKeyConfigWithDevEnvConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.novu_api_key.test", "value"),
					resource.TestCheckResourceAttrSet("data.novu_api_key.test", "owner_id"),
					resource.TestCheckResourceAttrSet("data.novu_api_key.test", "hash"),
				),
			},
		},
	})
}

var testApiKeyConfigWithDevEnvConfig = `
data "novu_environment" "test" {
	name = "Development"
}
data "novu_api_key" "test" {
	environment_id = data.novu_environment.test.id
}
`
