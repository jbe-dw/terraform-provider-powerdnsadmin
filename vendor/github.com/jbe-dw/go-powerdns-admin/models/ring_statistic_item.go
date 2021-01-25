// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// RingStatisticItem RingStatisticItem
//
// swagger:model RingStatisticItem
type RingStatisticItem struct {
	BaseStatisticItem

	RingStatisticItemAllOf1
}

// UnmarshalJSON unmarshals this object from a JSON structure
func (m *RingStatisticItem) UnmarshalJSON(raw []byte) error {
	// AO0
	var aO0 BaseStatisticItem
	if err := swag.ReadJSON(raw, &aO0); err != nil {
		return err
	}
	m.BaseStatisticItem = aO0

	// AO1
	var aO1 RingStatisticItemAllOf1
	if err := swag.ReadJSON(raw, &aO1); err != nil {
		return err
	}
	m.RingStatisticItemAllOf1 = aO1

	return nil
}

// MarshalJSON marshals this object to a JSON structure
func (m RingStatisticItem) MarshalJSON() ([]byte, error) {
	_parts := make([][]byte, 0, 2)

	aO0, err := swag.WriteJSON(m.BaseStatisticItem)
	if err != nil {
		return nil, err
	}
	_parts = append(_parts, aO0)

	aO1, err := swag.WriteJSON(m.RingStatisticItemAllOf1)
	if err != nil {
		return nil, err
	}
	_parts = append(_parts, aO1)
	return swag.ConcatJSON(_parts...), nil
}

// Validate validates this ring statistic item
func (m *RingStatisticItem) Validate(formats strfmt.Registry) error {
	var res []error

	// validation for a type composition with BaseStatisticItem
	if err := m.BaseStatisticItem.Validate(formats); err != nil {
		res = append(res, err)
	}
	// validation for a type composition with RingStatisticItemAllOf1
	if err := m.RingStatisticItemAllOf1.Validate(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// MarshalBinary interface implementation
func (m *RingStatisticItem) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *RingStatisticItem) UnmarshalBinary(b []byte) error {
	var res RingStatisticItem
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
