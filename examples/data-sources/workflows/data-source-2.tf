data "novu_workflows" "example2" {
  search = "example-workflow-name-or-other"
}

output "workflow" {
  value = data.novu_workflows.example2.items[0]
}
