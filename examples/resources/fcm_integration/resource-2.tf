locals {
  fcm_json_string = "{\"example\" : \"example\"}"
}

resource "novu_fcm_integration" "example2" {
  name               = "example fcm integration 2"
  identifier         = "example fcm integration 2"
  active             = true
  check              = true
  json_configuration = local.fcm_json_string
}
