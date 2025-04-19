package anytype

import (
	"fmt"
	"strings"
)

// extractTags is a helper function to extract tags from an object's Relations and Properties
func extractTags(obj *Object) {
	// Initialize Tags as an empty slice if nil
	if obj.Tags == nil {
		obj.Tags = []string{}
	}

	extractTagsFromRelations(obj)
	extractTagsFromProperties(obj)
}

// extractTagsFromRelations extracts tags from object's Relations field
func extractTagsFromRelations(obj *Object) {
	if obj.Relations == nil || obj.Relations.Items == nil {
		return
	}

	tagRelations, ok := obj.Relations.Items["tags"]
	if !ok || len(tagRelations) == 0 {
		return
	}

	// Extract name from each relation
	for _, relation := range tagRelations {
		if relation.Name != "" {
			obj.Tags = append(obj.Tags, relation.Name)
		}
	}
}

// extractTagsFromProperties extracts tags from object's Properties array
func extractTagsFromProperties(obj *Object) {
	if len(obj.Properties) == 0 {
		return
	}

	for _, prop := range obj.Properties {
		if !isTagProperty(prop) {
			continue
		}

		for _, tag := range prop.MultiSelect {
			if tag.Name != "" {
				obj.Tags = append(obj.Tags, tag.Name)
			}
		}
	}
}

// isTagProperty checks if a property is a tag property
func isTagProperty(prop Property) bool {
	return prop.Name == "Tag" &&
		prop.Format == "multi_select" &&
		len(prop.MultiSelect) > 0
}

// initializeTypeCache initializes the type cache for a specific space
func (c *Client) initializeTypeCache(spaceID string) {
	if _, ok := c.typeCache[spaceID]; !ok {
		c.typeCache[spaceID] = make(map[string]string)
	}
}

// buildReverseCache builds a reverse lookup (name -> key) from existing cache
func (c *Client) buildReverseCache(spaceID string) map[string]string {
	reverseCache := make(map[string]string)

	if cache, ok := c.typeCache[spaceID]; ok && len(cache) > 0 {
		for key, name := range cache {
			reverseCache[name] = key
		}
	}

	return reverseCache
}

// updateTypeCaches updates both the regular and reverse caches with type data
func (c *Client) updateTypeCaches(spaceID string, types []TypeInfo, typeName string) map[string]string {
	reverseCache := make(map[string]string)

	for _, t := range types {
		c.typeCache[spaceID][t.Key] = t.Name
		reverseCache[t.Name] = t.Key

		// Handle special case: "Page" -> "ot-page", "Note" -> "ot-note" etc.
		if c.isOtPrefixMatch(t.Key, typeName) {
			reverseCache[typeName] = t.Key
		}
	}

	return reverseCache
}

// isOtPrefixMatch checks if a key with "ot-" prefix matches the typeName
func (c *Client) isOtPrefixMatch(key, typeName string) bool {
	if strings.HasPrefix(key, "ot-") && strings.EqualFold(strings.TrimPrefix(key, "ot-"), typeName) {
		if c.debug && c.logger != nil {
			c.logger.Debug("Found matching type by key prefix: '%s' -> '%s'", typeName, key)
		}
		return true
	}
	return false
}

// findTypeKeyWithStrategies tries different strategies to find a type key
func (c *Client) findTypeKeyWithStrategies(typeName string, types []TypeInfo, reverseCache map[string]string) (string, error) {

	// Strategy 1: Exact match from updated cache
	if typeKey, found := reverseCache[typeName]; found {
		return typeKey, nil
	}

	// Strategy 2: Case-insensitive matching
	typeKey := c.findTypeKeyCaseInsensitive(typeName, reverseCache)
	if typeKey != "" {
		return typeKey, nil
	}

	// Strategy 3: Standard key construction (e.g., "Page" -> "ot-page")
	typeKey = c.findTypeKeyByStandardConstruction(typeName, types)
	if typeKey != "" {
		return typeKey, nil
	}

	return "", fmt.Errorf("type '%s' not found", typeName)
}

// findTypeKeyCaseInsensitive tries to find a type key using case-insensitive matching
func (c *Client) findTypeKeyCaseInsensitive(typeName string, reverseCache map[string]string) string {
	for name, key := range reverseCache {
		if strings.EqualFold(name, typeName) {
			if c.debug && c.logger != nil {
				c.logger.Debug("Found type using case-insensitive match: '%s' -> '%s'", typeName, name)
			}
			return key
		}
	}
	return ""
}

// findTypeKeyByStandardConstruction tries to construct a standard key format
func (c *Client) findTypeKeyByStandardConstruction(typeName string, types []TypeInfo) string {
	standardKey := "ot-" + strings.ToLower(typeName)
	for _, t := range types {
		if t.Key == standardKey {
			if c.debug && c.logger != nil {
				c.logger.Debug("Found type using standard key construction: '%s' -> '%s'", typeName, standardKey)
			}
			return standardKey
		}
	}
	return ""
}
