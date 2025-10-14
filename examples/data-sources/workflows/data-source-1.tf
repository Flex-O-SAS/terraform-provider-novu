data "novu_workflows" "example1" {
}

output "workflows" {
  value = data.novu_workflows.example1.items
}
