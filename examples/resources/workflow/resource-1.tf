resource "novu_fcm_integration" "example1" {
  name               = "example fcm integration"
  identifier         = "example fcm integration"
  active             = true
  check              = true
  json_configuration = "{\"example\" : \"example\"}"
}

resource "novu_workflow" "example1" {
  depends_on  = [novu_fcm_integration.example1] # To avoid issues in the push_step step
  workflow_id = "example-workflow-id"
  name        = "example workflow"
  steps = [
    {
      push_step = {
        name = "example push step"
        control_values = {
          subject = "example subject"
          body    = "example body"
        }
      }
    }
  ]
}
