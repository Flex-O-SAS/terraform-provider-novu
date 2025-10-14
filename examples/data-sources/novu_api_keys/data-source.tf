# Retrieve all the environment api keys
data "novu_api_keys" "example" {
}

locals {
  api_key = data.novu_api_keys.example.items[0]
}
