# Retrieve the api key for the specified environment
data "novu_environment" "example" {
  name = "Development"
}

data "novu_api_key" "example" {
  environment_id = data.novu_environment.example.id
}

locals {
  api_key = data.novu_api_key.example
}
