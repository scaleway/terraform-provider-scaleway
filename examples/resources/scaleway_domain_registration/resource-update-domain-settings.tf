### Update Domain Settings

resource "scaleway_domain_registration" "test" {
  domain_names      = ["example.com"]
  duration_in_years = 1

  owner_contact {
    legal_form                  = "individual"
    firstname                   = "John"
    lastname                    = "DOE"
    email                       = "john.doe@example.com"
    phone_number                = "+1.23456789"
    address_line_1              = "123 Main Street"
    city                        = "Paris"
    zip                         = "75001"
    country                     = "FR"
    vat_identification_code     = "FR12345678901"
    company_identification_code = "123456789"
  }

  auto_renew = true
  dnssec     = true
}
