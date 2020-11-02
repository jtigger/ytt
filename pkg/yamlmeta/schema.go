// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yamlmeta

import (
	"fmt"
)

type Schema interface {
	AssignType(typeable Typeable) TypeCheck
}

var _ Schema = &AnySchema{}
var _ Schema = &DocumentSchema{}

type AnySchema struct {
}

type DocumentSchema struct {
	Name    string
	Source  *Document
	Allowed *DocumentType
}

func NewDocumentSchema(doc *Document) (*DocumentSchema, error) {
	docType := &DocumentType{Source: doc}

	switch typedDocumentValue := doc.Value.(type) {
	case *Map:
		valueType, err := NewMapType(typedDocumentValue)
		if err != nil {
			return nil, err
		}

		docType.ValueType = valueType
	case *Array:
		valueType, err := NewArrayType(typedDocumentValue)
		if err != nil {
			return nil, err
		}

		docType.ValueType = valueType
	}
	return &DocumentSchema{
		Name:    "dataValues",
		Source:  doc,
		Allowed: docType,
	}, nil
}

func NewMapType(m *Map) (*MapType, error) {
	mapType := &MapType{}

	for _, mapItem := range m.Items {
		mapItemType, err := NewMapItemType(mapItem)
		if err != nil {
			return nil, err
		}
		mapType.Items = append(mapType.Items, mapItemType)
	}
	return mapType, nil
}

func NewMapItemType(item *MapItem) (*MapItemType, error) {
	switch typedContent := item.Value.(type) {
	case *Map:
		mapType, err := NewMapType(typedContent)
		if err != nil {
			return nil, err
		}
		return &MapItemType{Key: item.Key, ValueType: mapType}, nil
	case *Array:
		arrayType, err := NewArrayType(typedContent)
		if err != nil {
			return nil, err
		}
		return &MapItemType{Key: item.Key, ValueType: arrayType}, nil
	case string:
		return &MapItemType{Key: item.Key, ValueType: &ScalarType{Type: *new(string)}}, nil
	case int:
		return &MapItemType{Key: item.Key, ValueType: &ScalarType{Type: *new(int)}}, nil
	case bool:
		return &MapItemType{Key: item.Key, ValueType: &ScalarType{Type: *new(bool)}}, nil
	}
	return nil, fmt.Errorf("Map Item type did not match any known types")
}

func NewArrayType(a *Array) (*ArrayType, error) {
	if len(a.Items) != 1 {
		return nil, fmt.Errorf("Too many elements in the array to determine type. Need 1, given %n", len(a.Items))
	}

	switch typedContent := a.Items[0].Value.(type) {
	case *Map:
		mapType, err := NewMapType(typedContent)
		if err != nil {
			return nil, err
		}
		return &ArrayType{ItemsType: mapType}, nil
	case *Array:
		arrayType, err := NewArrayType(typedContent)
		if err != nil {
			return nil, err
		}
		return &ArrayType{ItemsType: arrayType}, nil
	case string:
		return &ArrayType{ItemsType: &ScalarType{Type: *new(string)}}, nil
	case int:
		return &ArrayType{ItemsType: &ScalarType{Type: *new(int)}}, nil
	case bool:
		return &ArrayType{ItemsType: &ScalarType{Type: *new(bool)}}, nil
	}


	return nil, fmt.Errorf("Array type did not match any know types")
}

func (as *AnySchema) AssignType(typeable Typeable) TypeCheck { return TypeCheck{} }

func (s *DocumentSchema) AssignType(typeable Typeable) TypeCheck {
	return s.Allowed.AssignTypeTo(typeable)
}
