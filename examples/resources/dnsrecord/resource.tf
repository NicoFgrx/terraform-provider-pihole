terraform {
  required_providers {
    pihole = {
      source = "localhost/dev/pihole"
    }
  }
}

# search env variables
provider "pihole" {
  url = "http://localhost:8080/admin/api.php"
  token = "96cf46f9e9312ea9ad00f5f9e63b25643f701246357068549a6c2ea3d163bf1e"
}

resource "pihole_dnsrecord" "example-1" {
  domain = "test1.example.com"
  ip     = "1.1.1.1"
}

resource "pihole_dnsrecord" "example-2" {
  domain = "test2.example.com"
  ip     = "2.2.2.2"
}
