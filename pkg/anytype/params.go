
package anytype

// GetObjectParams represents parameters for retrieving an object
type GetObjectParams struct {
	SpaceID  string `json:"space_id"`  // Space ID the object belongs to
	ObjectID string `json:"object_id"` // Object ID to retrieve
}

// Validate validates GetObjectParams fields
func (p *GetObjectParams) Validate() error {
	if p.SpaceID == "" {
		return ErrInvalidSpaceID
	}
	if p.ObjectID == "" {
		return ErrInvalidObjectID
	}
	return nil
}

// CreateObjectParams represents parameters for creating an object
type CreateObjectParams struct {
	SpaceID string  `json:"space_id"`         // Space ID where object should be created
	Object  *Object `json:"object,omitempty"` // Object to create
}

// Validate validates CreateObjectParams fields
func (p *CreateObjectParams) Validate() error {
	if p.SpaceID == "" {
		return ErrInvalidSpaceID
	}
	if p.Object == nil {
		return ErrInvalidParameter
	}
	return p.Object.Validate()
}

// UpdateObjectParams represents parameters for updating an object
type UpdateObjectParams struct {
	SpaceID  string  `json:"space_id"`         // Space ID the object belongs to
	ObjectID string  `json:"object_id"`        // Object ID to update
	Object   *Object `json:"object,omitempty"` // Object data for update
}

// Validate validates UpdateObjectParams fields
func (p *UpdateObjectParams) Validate() error {
	if p.SpaceID == "" {
		return ErrInvalidSpaceID
	}
	if p.ObjectID == "" {
		return ErrInvalidObjectID
	}
	if p.Object == nil {
		return ErrInvalidParameter
	}
	return nil
}

// DeleteObjectParams represents parameters for deleting an object
type DeleteObjectParams struct {
	SpaceID  string `json:"space_id"`  // Space ID the object belongs to
	ObjectID string `json:"object_id"` // Object ID to delete
}

// Validate validates DeleteObjectParams fields
func (p *DeleteObjectParams) Validate() error {
	if p.SpaceID == "" {
		return ErrInvalidSpaceID
	}
	if p.ObjectID == "" {
		return ErrInvalidObjectID
	}
	return nil
}

// GetSpacesParams represents parameters for retrieving spaces
type GetSpacesParams struct {
	IncludeMembers bool `json:"include_members,omitempty"` // Whether to include member information for each space
}

// NewGetSpacesParams creates a new GetSpacesParams with default values
func NewGetSpacesParams() *GetSpacesParams {
	return &GetSpacesParams{
		IncludeMembers: true,
	}
}

// GetSpaceByIDParams represents parameters for retrieving a specific space
type GetSpaceByIDParams struct {
	SpaceID string `json:"space_id"` // Space ID to retrieve
}

// Validate validates GetSpaceByIDParams fields
func (p *GetSpaceByIDParams) Validate() error {
	if p.SpaceID == "" {
		return ErrInvalidSpaceID
	}
	return nil
}

// Validate validates GetSpacesParams fields
func (p *GetSpacesParams) Validate() error {
	// No required fields to validate
	return nil
}

// GetTypesParams represents parameters for retrieving types from a space
type GetTypesParams struct {
	SpaceID string `json:"space_id"` // Space ID to get types from
}

// Validate validates GetTypesParams fields
func (p *GetTypesParams) Validate() error {
	if p.SpaceID == "" {
		return ErrInvalidSpaceID
	}
	return nil
}

// GetTypeByNameParams represents parameters for retrieving a type by name
type GetTypeByNameParams struct {
	SpaceID  string `json:"space_id"`  // Space ID to get type from
	TypeName string `json:"type_name"` // Type name to lookup
}

// Validate validates GetTypeByNameParams fields
func (p *GetTypeByNameParams) Validate() error {
	if p.SpaceID == "" {
		return ErrInvalidSpaceID
	}
	if p.TypeName == "" {
		return ErrInvalidTypeID
	}
	return nil
}

// GetMembersParams represents parameters for retrieving space members
type GetMembersParams struct {
	SpaceID string `json:"space_id"` // Space ID to get members from
}

// Validate validates GetMembersParams fields
func (p *GetMembersParams) Validate() error {
	if p.SpaceID == "" {
		return ErrInvalidSpaceID
	}
	return nil
}

// ExportImageParams represents parameters for exporting an image
type ExportImageParams struct {
	URL       string `json:"url"`        // The URL of the image to export
	OutputDir string `json:"output_dir"` // The directory to save the image to
}

// Validate validates ExportImageParams fields
func (p *ExportImageParams) Validate() error {
	if p.URL == "" {
		return ErrInvalidImageURL
	}
	return nil
}
