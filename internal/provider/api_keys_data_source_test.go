package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestApiKeysDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: `
				data "novu_api_keys" "test" {
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.novu_api_keys.test", "items.#", "2"),
					resource.TestCheckResourceAttrSet("data.novu_api_keys.test", "items.0.environment_id"),
					resource.TestCheckResourceAttrSet("data.novu_api_keys.test", "items.0.value"),
					resource.TestCheckResourceAttrSet("data.novu_api_keys.test", "items.0.hash"),
				),
			},
			// random id
			{
				Config: testApiKeysConfigWithRandomIdConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.novu_api_keys.test", "items.#", "0"),
				),
			},
			{
				Config: testApiKeysConfigWithEnvironmentConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.novu_api_keys.test", "items.#", "1"),
					resource.TestCheckResourceAttrSet("data.novu_api_keys.test", "items.0.environment_id"),
					resource.TestCheckResourceAttrSet("data.novu_api_keys.test", "items.0.value"),
					resource.TestCheckResourceAttrSet("data.novu_api_keys.test", "items.0.hash"),
				),
			},
		},
	})
}

var testApiKeysConfigWithRandomIdConfig = `
data "novu_api_keys" "test" {
	environment_id = "non-existing-id"
}
`

var testApiKeysConfigWithEnvironmentConfig = `
data "novu_environment" "test" {
	name = "Development"
}
data "novu_api_keys" "test" {
	environment_id = data.novu_environment.test.id
}
`
