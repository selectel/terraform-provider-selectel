terraform {
  required_providers {
    selectel = {
      source = "selectel/selectel"
      version = "~> 6.0.0"
    }
  }
}

# Create a project
resource "selectel_vpc_project_v2" "project_1" {
  ...
}
