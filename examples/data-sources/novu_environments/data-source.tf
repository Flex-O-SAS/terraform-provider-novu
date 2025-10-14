# List all the production environments in the organization
data "novu_environments" "example2" {
  is_production = true
}

