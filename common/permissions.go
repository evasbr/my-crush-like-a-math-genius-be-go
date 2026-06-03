// Package common provides cross-cutting utility helper functions
// that can be utilized across all layers of the application.
package common

import "strings"

// MergePermissions consolidates permissions from multiple sources (like different roles).
// It merges permission lists without duplicates and standardizes the "FULLACCESS" key to be uppercase.
func MergePermissions(permissionsList []map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	for _, permissions := range permissionsList {
		for k, v := range permissions {
			upperK := strings.ToUpper(k)
			if upperK == "FULLACCESS" {
				if valBool, ok := v.(bool); ok && valBool {
					merged["FULLACCESS"] = true
				}
				continue
			}

			var newPerms []string
			if slice, ok := v.([]interface{}); ok {
				for _, item := range slice {
					if itemStr, ok := item.(string); ok {
						newPerms = append(newPerms, itemStr)
					}
				}
			} else if sliceStr, ok := v.([]string); ok {
				newPerms = append(newPerms, sliceStr...)
			}

			if len(newPerms) > 0 {
				if existing, exists := merged[upperK]; exists {
					if existingSlice, ok := existing.([]string); ok {
						permMap := make(map[string]bool)
						for _, p := range existingSlice {
							permMap[p] = true
						}
						for _, p := range newPerms {
							permMap[p] = true
						}
						var finalSlice []string
						for p := range permMap {
							finalSlice = append(finalSlice, p)
						}
						merged[upperK] = finalSlice
					}
				} else {
					merged[upperK] = newPerms
				}
			}
		}
	}
	return merged
}
