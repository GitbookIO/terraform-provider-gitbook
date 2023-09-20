resource "gitbook_entity_schema" "example_entity_schema" {
  organization = "4Me7JapjYF3sgxrFoKxP" # Typically you would reference a variable
  type         = "terraform:example"    # Needs to be prefixed with `terraform:`
  title = {
    "singular" : "Example",
    "plural" : "Examples"
  }
  properties = [
    {
      "name" : "name",
      "title" : "Name",
      "type" : "text",
    },
    {
      "name" : "age",
      "title" : "Age",
      "type" : "number",
    },
    {
      "name" : "subscribed",
      "title" : "Subscribed",
      "type" : "boolean",
    }
  ]
}
