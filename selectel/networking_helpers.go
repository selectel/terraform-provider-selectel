package selectel

// Small conversion helpers shared by the networking resources. The SDK takes
// pointers for optional request fields and returns pointers for nullable
// response fields, while the Terraform SDK works with plain values.

// vpcString returns a pointer to value, for an optional request field.
func vpcString(value string) *string {
	return &value
}

// vpcBool returns a pointer to value, for an optional request field.
func vpcBool(value bool) *bool {
	return &value
}

// vpcInt returns a pointer to value, for an optional request field.
func vpcInt(value int) *int {
	return &value
}

// derefString returns the pointed-to string, or an empty string for a JSON null.
func derefString(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}

// derefInt returns the pointed-to int, or zero for a JSON null.
func derefInt(value *int) int {
	if value == nil {
		return 0
	}

	return *value
}

// networkingSubnetV2Object names the subnet entity in diagnostics; it is
// declared here so the network data source can reference it before the subnet
// resource lands.
const networkingSubnetV2Object = "networking subnet"
