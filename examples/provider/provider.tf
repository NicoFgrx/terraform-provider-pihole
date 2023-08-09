terraform {
  required_providers {
    pihole = {
      source = "localhost/dev/pihole"
    }
  }
}

provider "pihole" {}

resource "pihole_dnsrecord" "example" {}