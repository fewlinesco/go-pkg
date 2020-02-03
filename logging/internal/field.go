package internal

import (
	"strconv"
)

type IntField struct {
	Name  string
	Value int
}

func (i IntField) GetName() string  { return i.Name }
func (i IntField) GetValue() string { return strconv.Itoa(i.Value) }

type StringField struct {
	Name  string
	Value string
}

func (s StringField) GetName() string  { return s.Name }
func (s StringField) GetValue() string { return s.Value }
