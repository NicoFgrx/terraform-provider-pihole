---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "pihole_cname Resource - terraform-provider-pihole"
subcategory: ""
description: |-
  CNAME Record resource for pihole
---

# pihole_cname (Resource)

CNAME Record resource for pihole



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `domain` (String) Alias to use on CNAME
- `target` (String) Local managed DNS record

### Read-Only

- `last_updated` (String) Timestamp of the last Terraform update of the cname record.