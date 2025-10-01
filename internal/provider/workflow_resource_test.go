package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestWorkflowResource_Basic(t *testing.T) {

	randInt := acctest.RandIntRange(0, 1000)
	rname := fmt.Sprintf("tf-acc-%d", randInt)
	workflow_id := fmt.Sprintf("tf-acc-%d", randInt)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// first test : minimal config
				Config: testAccMinimalWorkflowResourceConfig(workflow_id, rname),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue("novu_workflow.test", tfjsonpath.New("name"), knownvalue.StringExact(rname)),
						plancheck.ExpectKnownValue("novu_workflow.test", tfjsonpath.New("workflow_id"), knownvalue.StringExact(workflow_id)),
						plancheck.ExpectUnknownValue("novu_workflow.test", tfjsonpath.New("slug")),
						plancheck.ExpectKnownValue("novu_workflow.test", tfjsonpath.New("description"), knownvalue.StringExact("")),
						plancheck.ExpectKnownValue("novu_workflow.test", tfjsonpath.New("active"), knownvalue.Bool(false)),
						plancheck.ExpectUnknownValue("novu_workflow.test", tfjsonpath.New("origin")),
						plancheck.ExpectUnknownValue("novu_workflow.test", tfjsonpath.New("status")),
						plancheck.ExpectKnownValue("novu_workflow.test", tfjsonpath.New("validate_payload"), knownvalue.Bool(true)),
						plancheck.ExpectUnknownValue("novu_workflow.test", tfjsonpath.New("created_at")),
						plancheck.ExpectUnknownValue("novu_workflow.test", tfjsonpath.New("updated_at")),
						plancheck.ExpectUnknownValue("novu_workflow.test", tfjsonpath.New("last_triggered_at")),
						plancheck.ExpectKnownValue("novu_workflow.test", tfjsonpath.New("tags"), knownvalue.ListSizeExact(0)),
						plancheck.ExpectKnownValue("novu_workflow.test", tfjsonpath.New("steps"), knownvalue.ListSizeExact(0)),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("novu_workflow.test", tfjsonpath.New("name"), knownvalue.StringExact(rname)),
					statecheck.ExpectKnownValue("novu_workflow.test", tfjsonpath.New("workflow_id"), knownvalue.StringExact(workflow_id)),
					statecheck.ExpectKnownValue("novu_workflow.test", tfjsonpath.New("slug"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue("novu_workflow.test", tfjsonpath.New("description"), knownvalue.StringExact("")),
					statecheck.ExpectKnownValue("novu_workflow.test", tfjsonpath.New("tags"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue("novu_workflow.test", tfjsonpath.New("steps"), knownvalue.ListSizeExact(0)),
				},
			},
			{
				// second test : full config
				Config: testAccFullWorkflowResourceConfig(workflow_id, rname),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue("novu_workflow.test", tfjsonpath.New("name"), knownvalue.StringExact(rname)),
						plancheck.ExpectKnownValue("novu_workflow.test", tfjsonpath.New("tags"), knownvalue.ListSizeExact(2)),
						plancheck.ExpectKnownValue("novu_workflow.test", tfjsonpath.New("tags"),
							knownvalue.ListExact([]knownvalue.Check{
								knownvalue.StringExact("test tag1"),
								knownvalue.StringExact("test tag2"),
							}),
						),
						plancheck.ExpectKnownValue("novu_workflow.test", tfjsonpath.New("steps"), knownvalue.ListSizeExact(2)),
						plancheck.ExpectUnknownValue("novu_workflow.test", tfjsonpath.New("steps").AtSliceIndex(0).AtMapKey("push_step").AtMapKey("step_id")),
						plancheck.ExpectUnknownValue("novu_workflow.test", tfjsonpath.New("steps").AtSliceIndex(0).AtMapKey("push_step").AtMapKey("slug")),
						plancheck.ExpectUnknownValue("novu_workflow.test", tfjsonpath.New("steps").AtSliceIndex(0).AtMapKey("push_step").AtMapKey("origin")),
						plancheck.ExpectUnknownValue("novu_workflow.test", tfjsonpath.New("steps").AtSliceIndex(0).AtMapKey("push_step").AtMapKey("issues")),
						plancheck.ExpectUnknownValue("novu_workflow.test", tfjsonpath.New("steps").AtSliceIndex(1).AtMapKey("push_step").AtMapKey("issues")),
						plancheck.ExpectUnknownValue("novu_workflow.test", tfjsonpath.New("steps").AtSliceIndex(1).AtMapKey("push_step").AtMapKey("step_id")),
						plancheck.ExpectUnknownValue("novu_workflow.test", tfjsonpath.New("steps").AtSliceIndex(1).AtMapKey("push_step").AtMapKey("slug")),
						plancheck.ExpectUnknownValue("novu_workflow.test", tfjsonpath.New("steps").AtSliceIndex(1).AtMapKey("push_step").AtMapKey("origin")),
						plancheck.ExpectKnownValue("novu_workflow.test", tfjsonpath.New("steps"),
							knownvalue.ListPartial(map[int]knownvalue.Check{
								0: knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"push_step": knownvalue.ObjectPartial(map[string]knownvalue.Check{
										"name": knownvalue.StringExact("test push step"),
										"control_values": knownvalue.ObjectExact(map[string]knownvalue.Check{
											"subject": knownvalue.StringExact("test subject"),
											"body":    knownvalue.StringExact("test body"),
										}),
									}),
								}),
								1: knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"push_step": knownvalue.ObjectPartial(map[string]knownvalue.Check{
										"name":           knownvalue.StringExact("test push step 2"),
										"control_values": knownvalue.Null(),
									}),
								}),
							}),
						),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("novu_workflow.test", tfjsonpath.New("name"), knownvalue.StringExact(rname)),
					statecheck.ExpectKnownValue("novu_workflow.test", tfjsonpath.New("steps"), knownvalue.ListSizeExact(2)),
					statecheck.ExpectKnownValue("novu_workflow.test", tfjsonpath.New("steps"),
						knownvalue.ListPartial(map[int]knownvalue.Check{
							0: knownvalue.ObjectPartial(map[string]knownvalue.Check{
								"push_step": knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"name": knownvalue.StringExact("test push step"),
								}),
							}),
							1: knownvalue.ObjectPartial(map[string]knownvalue.Check{
								"push_step": knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"name": knownvalue.StringExact("test push step 2"),
								}),
							}),
						}),
					),
					// no push integration => integration error
					statecheck.ExpectKnownValue("novu_workflow.test", tfjsonpath.New("steps").AtSliceIndex(0).AtMapKey("push_step").AtMapKey("issues"),
						knownvalue.ObjectExact(
							map[string]knownvalue.Check{
								"controls":    knownvalue.Null(),
								"integration": knownvalue.ListSizeExact(1),
							},
						),
					),
					// no push integration => integration error
					// no control values => 2 errors
					statecheck.ExpectKnownValue("novu_workflow.test", tfjsonpath.New("steps").AtSliceIndex(1).AtMapKey("push_step").AtMapKey("issues"),
						knownvalue.ObjectExact(
							map[string]knownvalue.Check{
								"controls":    knownvalue.ListSizeExact(2),
								"integration": knownvalue.ListSizeExact(1),
							},
						),
					),
				},
			},
			{
				// third test : same config, expect no changes
				Config: testAccFullWorkflowResourceConfig(workflow_id, rname),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccMinimalWorkflowResourceConfig(workflow_id string, name string) string {
	return fmt.Sprintf(`
resource "novu_workflow" "test" {
  workflow_id = "%s"
  name = "%s"
}
`, workflow_id, name)
}

func testAccFullWorkflowResourceConfig(workflow_id string, name string) string {
	return fmt.Sprintf(`
resource "novu_workflow" "test" {
  workflow_id = "%s"
  name = "%s"
  description = "test description"
  active = true
  validate_payload = true
  tags = ["test tag1", "test tag2"]
  steps = [
    {
      push_step = {
        name = "test push step"
		control_values = {
		  subject = "test subject"
		  body = "test body"
		}
	  }
    },
	{
	  push_step = {
  		name = "test push step 2"
	  }
	}
  ]
}
`, workflow_id, name)
}
