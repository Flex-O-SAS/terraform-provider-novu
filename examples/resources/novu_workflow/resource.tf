resource "novu_fcm_integration" "example" {
  name               = "example fcm integration"
  identifier         = "example fcm integration"
  active             = true
  check              = true
  json_configuration = "{\"example\" : \"example\"}"
}

resource "novu_workflow" "example" {
  depends_on  = [novu_fcm_integration.example1] # To avoid integration issues in the push_step steps
  workflow_id = "example-workflow-id"
  name        = "example workflow"
  steps = [
    {
      push_step = {
        name = "first push step"
        control_values = {
          subject = "example subject"
          body    = "example body"
        }
      }
    },
    {
      push_step = {
        name = "second push step"
      }
    }
  ]
}

# Retrieve the issues for the workflow steps, if any
locals {
  integration_issues = flatten([
    for k, v in novu_workflow.example2.steps :
    try([v.push_step.issues.integration], [])
  ])
  control_issues = flatten([
    for k, v in novu_workflow.example2.steps :
    try([v.push_step.issues.controls], [])
  ])
}
