resource "novu_fcm_integration" "example1" {
  name       = "example fcm integration 1"
  identifier = "example fcm integration 1"
  active     = true
  check      = true
  service_account = {
    type                        = "service_account"
    project_id                  = "example-id"
    private_key_id              = "1234567890abcdef1234567890abcdef"
    private_key                 = "-----BEGIN PRIVATE KEY----- \n 1234567890abcdef1234567890abcdef \n -----END PRIVATE KEY-----"
    client_email                = "example@example.com"
    client_id                   = "123456"
    auth_uri                    = "https://example.com/auth"
    token_uri                   = "https://example.com/token"
    auth_provider_x509_cert_url = "https://example.com/auth-provider-x509-cert-url"
    client_x509_cert_url        = "https://example.com/client-x509-cert-url"
  }
}

# Alternatively, use the json_configuration : 
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
