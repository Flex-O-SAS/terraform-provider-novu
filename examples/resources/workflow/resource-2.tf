resource "novu_workflow" "example2" {
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
