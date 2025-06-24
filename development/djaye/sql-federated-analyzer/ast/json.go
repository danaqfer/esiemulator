package ast

import (
	"encoding/json"
	"fmt"
)

// MarshalJSON implements custom JSON marshaling for Expression interface
func (b *BinaryExpression) MarshalJSON() ([]byte, error) {
	type Alias BinaryExpression
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  "binary_expression",
		Alias: (*Alias)(b),
	})
}

// MarshalJSON implements custom JSON marshaling for Expression interface
func (c *ColumnReference) MarshalJSON() ([]byte, error) {
	type Alias ColumnReference
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  "column_reference",
		Alias: (*Alias)(c),
	})
}

// MarshalJSON implements custom JSON marshaling for Expression interface
func (l *Literal) MarshalJSON() ([]byte, error) {
	type Alias Literal
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  "literal",
		Alias: (*Alias)(l),
	})
}

// MarshalJSON implements custom JSON marshaling for JoinType
func (j JoinType) MarshalJSON() ([]byte, error) {
	var s string
	switch j {
	case INNER:
		s = "INNER"
	case LEFT:
		s = "LEFT"
	case RIGHT:
		s = "RIGHT"
	case FULL:
		s = "FULL"
	case CROSS:
		s = "CROSS"
	default:
		s = "UNKNOWN"
	}
	return json.Marshal(s)
}

// MarshalJSON implements custom JSON marshaling for LiteralType
func (lt LiteralType) MarshalJSON() ([]byte, error) {
	var s string
	switch lt {
	case STRING:
		s = "STRING"
	case NUMBER:
		s = "NUMBER"
	case BOOLEAN:
		s = "BOOLEAN"
	case NULL:
		s = "NULL"
	case DATE:
		s = "DATE"
	case TIMESTAMP:
		s = "TIMESTAMP"
	default:
		s = "UNKNOWN"
	}
	return json.Marshal(s)
}

// MarshalJSON implements custom JSON marshaling for Expression interface
func (f *FunctionCall) MarshalJSON() ([]byte, error) {
	type Alias FunctionCall
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  "function_call",
		Alias: (*Alias)(f),
	})
}

// MarshalJSON implements custom JSON marshaling for Expression interface
func (s *Star) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Type string `json:"type"`
	}{
		Type: "star",
	})
}

// ToJSON converts an AST node to a JSON string
func ToJSON(node Node) (string, error) {
	bytes, err := json.MarshalIndent(node, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal AST to JSON: %v", err)
	}
	return string(bytes), nil
}
