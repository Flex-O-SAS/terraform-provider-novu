package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestFCMIntegrationResourceBasic(t *testing.T) {
	randInt := acctest.RandIntRange(0, 1000)
	rname := fmt.Sprintf("tf-acc-%d", randInt)
	identifier := fmt.Sprintf("tf-acc-id-%d", randInt)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFCMIntegrationResourceConfigWithJsonConfiguration(rname, identifier),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("novu_fcm_integration.test", "name", rname),
					resource.TestCheckResourceAttr("novu_fcm_integration.test", "identifier", identifier),
					resource.TestCheckResourceAttr("novu_fcm_integration.test", "active", "true"),
					resource.TestCheckResourceAttr("novu_fcm_integration.test", "check", "true"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("novu_fcm_integration.test", tfjsonpath.New("json_configuration"), knownvalue.StringExact("{\"test\" : \"test\"}")),
					statecheck.ExpectKnownValue("novu_fcm_integration.test", tfjsonpath.New("id"), knownvalue.NotNull()),
				},
			},
			{
				Config: testAccFCMIntegrationResourceConfigWithServiceAccount(rname, identifier),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
					},
				},
			},
			{
				Config: testAccFCMIntegrationResourceConfigWithServiceAccount(rname, identifier),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestFCMIntegrationResourceErrors(t *testing.T) {
	randInt := acctest.RandIntRange(0, 1000)
	rname := fmt.Sprintf("tf-acc-err-%d", randInt)
	identifier := fmt.Sprintf("tf-acc-id-err-%d", randInt)

	t.Parallel()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccFCMIntegrationResourceConfigErrorIncomplete(rname, identifier),
				ExpectError: regexp.MustCompile(".*"),
			},
			{
				Config:      testAccFCMIntegrationResourceConfigErrorWithCheck(rname, identifier),
				ExpectError: regexp.MustCompile(".*"),
			},
			{
				Config:      testAccFCMIntegrationResourceConfigErrorWithAllFields(rname, identifier),
				ExpectError: regexp.MustCompile(".*"),
			},
			{
				Config:      testAccFCMIntegrationResourceConfigErrorWithPartialServiceAccount(rname, identifier),
				ExpectError: regexp.MustCompile(".*"),
			},
		},
	})
}

func testAccFCMIntegrationResourceConfigWithServiceAccount(name, identifier string) string {
	return fmt.Sprintf(`
resource "novu_fcm_integration" "test" {
	name = "%s"
	identifier = "%s"
	check = true
	active = true
	service_account = {
		type = "service_account"
		project_id = "test"
		private_key_id = "test"
		private_key = "test"
		client_email = "test"
		client_id = "test"
		auth_uri = "test"
		token_uri = "test"
		auth_provider_x509_cert_url = "test"
		client_x509_cert_url = "test"
	}
}
`, name, identifier)
}

func testAccFCMIntegrationResourceConfigWithJsonConfiguration(name, identifier string) string {
	return fmt.Sprintf(`
resource "novu_fcm_integration" "test" {
	name = "%s"
	identifier = "%s"
	check = true
	active = true
	json_configuration = "{\"test\" : \"test\"}"
}
`, name, identifier)
}

func testAccFCMIntegrationResourceConfigErrorIncomplete(name, identifier string) string {
	return fmt.Sprintf(`
resource "novu_fcm_integration" "test" {
	name = "%s"
	identifier = "%s"
}
`, name, identifier)
}

func testAccFCMIntegrationResourceConfigErrorWithCheck(name, identifier string) string {
	return fmt.Sprintf(`
resource "novu_fcm_integration" "test" {
	name = "%s"
	identifier = "%s"
	check = true
}
`, name, identifier)
}

func testAccFCMIntegrationResourceConfigErrorWithAllFields(name, identifier string) string {
	return fmt.Sprintf(`
resource "novu_fcm_integration" "test" {
	name = "%s"
	identifier = "%s"
	active = true
	check = true
	json_configuration = "{\"test\" : \"test\"}"
	service_account = {
		type = "service_account"
		project_id = "test"
		private_key_id = "test"
		private_key = "test"
		client_email = "test"
		client_id = "test"
		auth_uri = "test"
		token_uri = "test"
		auth_provider_x509_cert_url = "test"
		client_x509_cert_url = "test"
	}
}
`, name, identifier)
}

func testAccFCMIntegrationResourceConfigErrorWithPartialServiceAccount(name, identifier string) string {
	return fmt.Sprintf(`
resource "novu_fcm_integration" "test" {
	name = "%s"
	identifier = "%s"
	service_account = {
		type = "service_account"
		project_id = "test"
		private_key_id = "test"
		private_key = "test"
		client_email = "test"
		client_id = "test"
	}
}
`, name, identifier)
}
