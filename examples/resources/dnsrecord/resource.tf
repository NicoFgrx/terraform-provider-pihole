terraform {
  required_providers {
    pihole = {
      source = "localhost/dev/pihole"
    }
  }
}

# search env variables
provider "pihole" {}

resource "pihole_dnsrecord" "box" {
  domain = "box.pasfastoche.lan"
  ip     = "192.168.1.1"
}
