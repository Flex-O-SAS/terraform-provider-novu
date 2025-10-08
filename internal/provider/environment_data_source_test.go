package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestEnvironmentDataSource(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: `
				data "novu_environment" "test" {
				}
				`,
				ExpectError: regexp.MustCompile("No criteria provided"),
			},
			{
				Config: `
				data "novu_environment" "test" {
					id = "non-existing-id"
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.novu_environment.test", "error", "no environment found"),
				),
				ExpectError: regexp.MustCompile("No environment found"),
			},
			{
				Config: `
				data "novu_environment" "test" {
					name = "Development"
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.novu_environment.test", "name", "Development"),
					resource.TestCheckResourceAttrSet("data.novu_environment.test", "identifier"),
					resource.TestCheckResourceAttrSet("data.novu_environment.test", "slug"),
					resource.TestCheckResourceAttr("data.novu_environment.test", "is_production", "false"),
				),
			},
		},
	})
}
