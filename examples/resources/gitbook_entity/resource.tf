resource "gitbook_entity" "example" {
  organization_id = "4Me7JapjYF3sgxrFoKxP" # Typically you would reference a variable
  type            = "terraform:example"
  entity_id       = "example-id"
  properties = {
    "name" = {
      "string" = "Alice"
    },
    "age" = {
      "number" = 42
    },
    "subscribed" = {
      "boolean" = true
    }
  }
}
