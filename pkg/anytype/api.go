package anytype

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// GetSpaces retrieves all available spaces from the Anytype API.
//
// This method fetches all spaces that the authenticated user has access to.
// For each space, it also attempts to fetch and populate the space's members.
// If member fetching fails for any space, the error is logged (in debug mode) but
// the space is still included in the results.
//
// Example:
//
//	spaces, err := client.GetSpaces(ctx)
//	if err != nil {
//	    log.Fatalf("Failed to get spaces: %v", err)
//	}
//
//	fmt.Printf("Found %d spaces:\n", len(spaces.Data))
//	for _, space := range spaces.Data {
//	    fmt.Printf("- %s (ID: %s)\n", space.Name, space.ID)
//	}
func (c *Client) GetSpaces(ctx context.Context) (*SpacesResponse, error) {
	data, err := c.makeRequest(ctx, http.MethodGet, "/v1/spaces", nil)
	if err != nil {
		return nil, wrapError("/v1/spaces", 0, "failed to get spaces", err)
	}

	if c.debug && c.logger != nil {
		c.logger.Debug("Raw spaces response: %s", string(data))
	}

	var response SpacesResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, wrapError("/v1/spaces", 0, "failed to parse spaces response", err)
	}

	// Fetch members for each space
	for i := range response.Data {
		if c.debug && c.logger != nil {
			c.logger.Debug("Fetching members for space %s (%s)", response.Data[i].Name, response.Data[i].ID)
		}

		members, err := c.GetMembers(ctx, &GetMembersParams{SpaceID: response.Data[i].ID})
		if err != nil {
			if c.debug && c.logger != nil {
				c.logger.Debug("Warning: failed to get members for space %s: %v", response.Data[i].ID, err)
			}
			continue
		}

		response.Data[i].Members = members.Data
	}

	return &response, nil
}

// GetSpaceByID retrieves a specific space by ID
func (c *Client) GetSpaceByID(ctx context.Context, params *GetSpaceByIDParams) (*Space, error) {
	if params.SpaceID == "" {
		return nil, wrapError("/v1/spaces/{id}", 0, "space ID is required", ErrInvalidSpaceID)
	}

	path := fmt.Sprintf("/v1/spaces/%s", params.SpaceID)
	data, err := c.makeRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, wrapError(path, 0, fmt.Sprintf("failed to get space %s", params.SpaceID), err)
	}

	if c.debug && c.logger != nil {
		c.logger.Debug("Raw space response: %s", string(data))
	}

	var response struct {
		Space Space `json:"space"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, wrapError(path, 0, "failed to parse space response", err)
	}

	return &response.Space, nil
}

// GetMembers retrieves members of a space
func (c *Client) GetMembers(ctx context.Context, params *GetMembersParams) (*MembersResponse, error) {
	if params == nil {
		return nil, ErrInvalidParameter
	}
	if err := params.Validate(); err != nil {
		return nil, err
	}
	if params.SpaceID == "" {
		return nil, wrapError("/v1/spaces/{id}/members", 0, "space ID is required", ErrInvalidSpaceID)
	}

	path := fmt.Sprintf("/v1/spaces/%s/members", params.SpaceID)
	data, err := c.makeRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, wrapError(path, 0, fmt.Sprintf("failed to get members for space %s", params.SpaceID), err)
	}

	if c.debug && c.logger != nil {
		c.logger.Debug("Raw members response: %s", string(data))
	}

	var response MembersResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, wrapError(path, 0, "failed to parse members response", err)
	}

	return &response, nil
}

// GetTypes retrieves types from a space
func (c *Client) GetTypes(ctx context.Context, params *GetTypesParams) (*TypeResponse, error) {
	if params == nil {
		return nil, ErrInvalidParameter
	}
	if err := params.Validate(); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/v1/spaces/%s/types", params.SpaceID)
	data, err := c.makeRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get types for space %s: %w", params.SpaceID, err)
	}

	// This follows the API's pagination response format for types
	var response TypeResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse types response: %w", err)
	}

	// Update the type cache with the retrieved types
	// Initialize cache for this space if needed
	if _, ok := c.typeCache[params.SpaceID]; !ok {
		c.typeCache[params.SpaceID] = make(map[string]string)
	}

	// Update cache with all types
	for _, t := range response.Data {
		c.typeCache[params.SpaceID][t.Key] = t.Name
	}

	return &response, nil
}

// GetTypeByName retrieves key for a specific type by its name
func (c *Client) GetTypeByName(ctx context.Context, params *GetTypeByNameParams) (string, error) {
	if params.SpaceID == "" {
		return "", ErrInvalidSpaceID
	}
	if params.TypeName == "" {
		return "", ErrInvalidTypeID
	}

	// Initialize cache for this space if needed
	c.initializeTypeCache(params.SpaceID)

	// First try to find from the existing cache
	reverseCache := c.buildReverseCache(params.SpaceID)
	if typeKey, found := reverseCache[params.TypeName]; found {
		return typeKey, nil
	}

	// If not in cache, fetch all types and update cache
	types, err := c.GetTypes(ctx, &GetTypesParams{SpaceID: params.SpaceID})
	if err != nil {
		return "", err
	}

	// Update cache with fresh data
	reverseCache = c.updateTypeCaches(params.SpaceID, types.Data, params.TypeName)

	// Try different matching strategies in order
	return c.findTypeKeyWithStrategies(params.TypeName, types.Data, reverseCache)
}

// GetObject retrieves a specific object by ID.
//
// This method fetches an object from Anytype by its unique ID. The object includes
// metadata such as name, type, icon, tags, and other properties defined in the Object struct.
//
// The method requires a valid context and GetObjectParams struct containing the space ID
// and object ID. If successful, it returns a populated Object struct and nil error.
//
// If the object doesn't exist or cannot be fetched due to permissions or network issues,
// an appropriate error will be returned.
//
// Example:
//
//	params := &anytype.GetObjectParams{
//	    SpaceID:  "space123",
//	    ObjectID: "obj456",
//	}
//
//	object, err := client.GetObject(ctx, params)
//	if err != nil {
//	    log.Fatalf("Failed to get object: %v", err)
//	}
//
//	fmt.Printf("Object name: %s\n", object.Name)
func (c *Client) GetObject(ctx context.Context, params *GetObjectParams) (*Object, error) {
	if params == nil {
		return nil, ErrInvalidParameter
	}
	if err := params.Validate(); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/v1/spaces/%s/objects/%s", params.SpaceID, params.ObjectID)
	data, err := c.makeRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get object %s: %w", params.ObjectID, err)
	}

	// The API response is structured with an "object" field
	// containing the Object data in the standard response format
	var objectResponse struct {
		Object Object `json:"object"`
	}

	if err := json.Unmarshal(data, &objectResponse); err != nil {
		return nil, fmt.Errorf("failed to parse object response: %w", err)
	}

	// Extract tags from the retrieved object
	extractTags(&objectResponse.Object)

	return &objectResponse.Object, nil
}

// CreateObject creates a new object in a space.
//
// This method allows you to create new objects in a specified Anytype space.
// The object parameter must contain all required fields for the object type,
// such as name, type key, and any properties specific to that type.
//
// If the object contains tags in the Tags field, they will automatically be
// added to the object's Relations.
//
// Example:
//
//	// Create a new note
//	newObject := &anytype.Object{
//	    Name:    "Meeting Notes",
//	    TypeKey: "ot-note",
//	    Icon: &anytype.Icon{
//	        Format: "emoji",
//	        Emoji:  "📝",
//	    },
//	    Tags: []string{"work", "meeting"},
//	}
//
//	created, err := client.CreateObject(ctx, "space123", newObject)
//	if err != nil {
//	    log.Fatalf("Failed to create object: %v", err)
//	}
//
//	fmt.Printf("Created object with ID: %s\n", created.ID)
func (c *Client) CreateObject(ctx context.Context, params *CreateObjectParams) (*Object, error) {
	if params.SpaceID == "" {
		return nil, ErrInvalidSpaceID
	}
	if params.Object == nil {
		return nil, ErrInvalidParameter
	}

	// For object creation, we need to validate differently
	// The ID should be empty (server will assign it), but other validations still apply
	if params.Object.Type == nil || params.Object.Type.Key == "" {
		return nil, ErrInvalidTypeID
	}

	// Add a temporary ID if needed for passing the validation stage
	if params.Object.ID == "" {
		params.Object.ID = "temp-id-for-creation"
	}

	// Ensure we add tags to Relations if they're specified in the Tags field
	if len(params.Object.Tags) > 0 {
		if params.Object.Relations == nil {
			params.Object.Relations = &Relations{
				Items: make(map[string][]Relation),
			}
		}

		// Create tag relations from tag names
		tagRelations := make([]Relation, 0, len(params.Object.Tags))
		for _, tagName := range params.Object.Tags {
			tagRelations = append(tagRelations, Relation{
				Name: tagName,
			})
		}

		// Add to the relations map
		params.Object.Relations.Items["tags"] = tagRelations
	}

	path := fmt.Sprintf("/v1/spaces/%s/objects", params.SpaceID)

	// Always remove the ID for object creation requests, regardless of its value
	// The server will assign a proper ID to the new object
	params.Object.ID = ""

	// Create a proper request object according to API schema
	// Using fields that match the object.CreateObjectRequest schema in the Swagger spec
	createRequest := map[string]interface{}{
		"name": params.Object.Name,
	}

	// Add type_key - either from the TypeKey field or from the Type object
	if params.Object.TypeKey != "" {
		createRequest["type_key"] = params.Object.TypeKey
	} else if params.Object.Type != nil && params.Object.Type.Key != "" {
		createRequest["type_key"] = params.Object.Type.Key
	}

	// Add icon if present
	if params.Object.Icon != nil {
		createRequest["icon"] = params.Object.Icon
	}

	// Add tags if present - properly format them as the API expects
	if len(params.Object.Tags) > 0 {
		// Convert simple string tags to proper tag objects
		tagObjects := make([]map[string]string, 0, len(params.Object.Tags))
		for _, tagName := range params.Object.Tags {
			tagObjects = append(tagObjects, map[string]string{
				"name": tagName,
			})
		}
		createRequest["tags"] = tagObjects
	}

	// Add description - either from Description field or fall back to Snippet
	if params.Object.Description != "" {
		createRequest["description"] = params.Object.Description
	} else if params.Object.Snippet != "" {
		createRequest["description"] = params.Object.Snippet
	}

	// Add body content if present
	if params.Object.Body != "" {
		createRequest["body"] = params.Object.Body
	}

	// Add source URL for bookmarks
	if params.Object.Source != "" {
		createRequest["source"] = params.Object.Source
	}

	// Add template_id if specified
	if params.Object.TemplateID != "" {
		createRequest["template_id"] = params.Object.TemplateID
	}

	body, err := json.Marshal(createRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal object: %w", err)
	}

	data, err := c.makeRequest(ctx, http.MethodPost, path, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create object: %w", err)
	}

	// The API response is structured with an "object" field
	// containing the Object data in the standard response format
	var objectResponse struct {
		Object Object `json:"object"`
	}

	if err := json.Unmarshal(data, &objectResponse); err != nil {
		return nil, fmt.Errorf("failed to parse created object response: %w", err)
	}

	// Extract tags using the helper function
	extractTags(&objectResponse.Object)

	return &objectResponse.Object, nil
}

// DeleteObject deletes an object from a space.
//
// This method permanently removes an object identified by its params.ObjectID from the
// specified space. Once deleted, the object cannot be recovered through the API.
//
// Example:
//
//	err := client.DeleteObject(ctx, "space123", "obj456")
//	if err != nil {
//	    log.Fatalf("Failed to delete object: %v", err)
//	}
//
//	fmt.Println("Object successfully deleted")
func (c *Client) DeleteObject(ctx context.Context, params *DeleteObjectParams) error {
	if params.SpaceID == "" {
		return ErrInvalidSpaceID
	}
	if params.ObjectID == "" {
		return ErrInvalidObjectID
	}

	path := fmt.Sprintf("/v1/spaces/%s/objects/%s", params.SpaceID, params.ObjectID)
	_, err := c.makeRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("failed to delete object %s: %w", params.ObjectID, err)
	}

	return nil
}

// UpdateObject updates an existing object in a space.
//
// This method allows you to update an existing object identified by its params.ObjectID
// within the specified space. The provided object parameter should contain the
// fields to be updated. The object ID in the object parameter will be overridden
// with the params.ObjectID parameter to ensure consistency.
//
// If the object contains tags in the Tags field, they will automatically be
// added to the object's Relations, replacing any existing tag relations.
//
// Example:
//
//	// Update an existing object
//	updateObj := &anytype.Object{
//	    Name: "Updated Meeting Notes",
//	    Icon: &anytype.Icon{
//	        Format: "emoji",
//	        Emoji:  "📌",
//	    },
//	    Tags: []string{"work", "important", "meeting"},
//	}
//
//	updated, err := client.UpdateObject(ctx, "space123", "obj456", updateObj)
//	if err != nil {
//	    log.Fatalf("Failed to update object: %v", err)
//	}
//
//	fmt.Printf("Updated object: %s\n", updated.Name)
func (c *Client) UpdateObject(ctx context.Context, params *UpdateObjectParams) (*Object, error) {
	if params.SpaceID == "" {
		return nil, ErrInvalidSpaceID
	}
	if params.ObjectID == "" {
		return nil, ErrInvalidObjectID
	}
	if params.Object == nil {
		return nil, ErrInvalidParameter
	}

	// Ensure the object ID in the path matches the object ID in the body
	params.Object.ID = params.ObjectID

	// Ensure we add tags to Relations if they're specified in the Tags field
	if len(params.Object.Tags) > 0 {
		if params.Object.Relations == nil {
			params.Object.Relations = &Relations{
				Items: make(map[string][]Relation),
			}
		}

		// Create tag relations from tag names
		tagRelations := make([]Relation, 0, len(params.Object.Tags))
		for _, tagName := range params.Object.Tags {
			tagRelations = append(tagRelations, Relation{
				Name: tagName,
			})
		}

		// Add to the relations map
		params.Object.Relations.Items["tags"] = tagRelations
	}

	// For updates, use the objects endpoint without the object ID - the ID will be in the request body
	path := fmt.Sprintf("/v1/spaces/%s/objects", params.SpaceID)

	// Create a proper update request similar to the create request format
	updateRequest := map[string]interface{}{
		"id":   params.ObjectID, // Include the object ID so the API knows which object to update
		"name": params.Object.Name,
	}

	// Add type_key if present
	if params.Object.TypeKey != "" {
		updateRequest["type_key"] = params.Object.TypeKey
	} else if params.Object.Type != nil && params.Object.Type.Key != "" {
		updateRequest["type_key"] = params.Object.Type.Key
	}

	// Add icon if present
	if params.Object.Icon != nil {
		updateRequest["icon"] = params.Object.Icon
	}

	// Add tags if present - properly format them as the API expects
	if len(params.Object.Tags) > 0 {
		// Convert simple string tags to proper tag objects
		tagObjects := make([]map[string]string, 0, len(params.Object.Tags))
		for _, tagName := range params.Object.Tags {
			tagObjects = append(tagObjects, map[string]string{
				"name": tagName,
			})
		}
		updateRequest["tags"] = tagObjects
	}

	// Add description if present
	if params.Object.Description != "" {
		updateRequest["description"] = params.Object.Description
	} else if params.Object.Snippet != "" {
		updateRequest["description"] = params.Object.Snippet
	}

	body, err := json.Marshal(updateRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal object: %w", err)
	}

	// Using POST instead of PATCH, as the Anytype API doesn't support PATCH
	// The same create endpoint is used for updates, but we include the object ID
	data, err := c.makeRequest(ctx, http.MethodPost, path, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to update object %s: %w", params.ObjectID, err)
	}

	// The API response is structured with an "object" field
	// containing the Object data in the standard response format
	var objectResponse struct {
		Object Object `json:"object"`
	}

	if err := json.Unmarshal(data, &objectResponse); err != nil {
		return nil, fmt.Errorf("failed to parse updated object response: %w", err)
	}

	// Extract tags using the helper function
	extractTags(&objectResponse.Object)

	return &objectResponse.Object, nil
}
