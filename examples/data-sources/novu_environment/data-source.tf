data "novu_environment" "example-development" {
  is_production = false
  name          = "Development"
}

data "novu_environment" "example-child-environment" {
  parent_environment_id = data.novu_environment.example-development.id
}
