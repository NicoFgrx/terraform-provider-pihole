---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "pihole Provider"
subcategory: ""
description: |-
  Interact with Pihole.
---

# pihole Provider

Interact with Pihole.

## Example Usage

```terraform
terraform {
  required_providers {
    pihole = {
      source = "localhost/dev/pihole"
    }
  }
}

provider "pihole" {}

resource "pihole_dnsrecord" "example" {}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `token` (String, Sensitive) Token for Pihole API. May also be provided via PIHOLE_TOKEN environment variable.
- `url` (String) URI for Pihole API. May also be provided via PIHOLE_API_URL environment variable.