package utils

import "reflect"

type AgentPrompts struct {
	Role      string
	Goal      string
	Backstory string
}

func GetRequiredFields(v interface{}) []string {
	var requiredFields []string

	// Get the reflect Type and Value of the input struct
	t := reflect.TypeOf(v)

	// Iterate over the struct fields
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		requiredTag := field.Tag.Get("required")

		// Check if the field has the required:"true" tag
		if requiredTag == "true" {
			jsonTag := field.Tag.Get("json")

			// Append the json tag value to the requiredFields slice
			requiredFields = append(requiredFields, jsonTag)
		}
	}

	return requiredFields
}
