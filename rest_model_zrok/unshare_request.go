// Code generated by go-swagger; DO NOT EDIT.

package rest_model_zrok

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// UnshareRequest unshare request
//
// swagger:model unshareRequest
type UnshareRequest struct {

	// env z Id
	EnvZID string `json:"envZId,omitempty"`

	// reserved
	Reserved bool `json:"reserved,omitempty"`

	// svc token
	SvcToken string `json:"svcToken,omitempty"`
}

// Validate validates this unshare request
func (m *UnshareRequest) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this unshare request based on context it is used
func (m *UnshareRequest) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *UnshareRequest) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *UnshareRequest) UnmarshalBinary(b []byte) error {
	var res UnshareRequest
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
