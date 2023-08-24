---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "gitbook_entity_schema Data Source - terraform-provider-gitbook"
subcategory: ""
description: |-
  Entity schema data source
---

# gitbook_entity_schema (Data Source)

Entity schema data source



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `properties` (Attributes Set) (see [below for nested schema](#nestedatt--properties))
- `title` (String)
- `type` (String)

<a id="nestedatt--properties"></a>
### Nested Schema for `properties`

Required:

- `description` (String)
- `name` (String)
- `title` (Attributes) (see [below for nested schema](#nestedatt--properties--title))
- `type` (String)

Optional:

- `entity` (Attributes) (see [below for nested schema](#nestedatt--properties--entity))

<a id="nestedatt--properties--title"></a>
### Nested Schema for `properties.title`

Required:

- `plural` (String)
- `singular` (String)


<a id="nestedatt--properties--entity"></a>
### Nested Schema for `properties.entity`

Required:

- `type` (String)

Optional:

- `integration` (String)