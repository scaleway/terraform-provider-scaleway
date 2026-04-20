data "scaleway_interlink_pop" "pop" {
  name = "Telehouse TH2"
}

data "scaleway_interlink_partner" "partner" {
  name = "FranceIX"
}

resource "scaleway_vpc" "vpc" {
  name = "my-vpc"
}

resource "scaleway_interlink_link" "main" {
  name           = "my-hosted-link"
  pop_id         = data.scaleway_interlink_pop.pop.id
  partner_id     = data.scaleway_interlink_partner.partner.id
  bandwidth_mbps = 50
  vpc_id         = scaleway_vpc.vpc.id
}
