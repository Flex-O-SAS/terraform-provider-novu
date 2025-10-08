data "novu_environment" "example" {
  name = "Development"
}
data "novu_api_keys" "example" {
  environment_id = data.novu_environment.example.id
}
