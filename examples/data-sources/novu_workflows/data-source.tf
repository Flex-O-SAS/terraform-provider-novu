# Retrieve a workflow by name
data "novu_workflows" "example2" {
  search = "example-workflow-name"
}

output "workflow" {
  value = data.novu_workflows.example2.items[0]
}
