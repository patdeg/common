package common

import "testing"

func TestAddIfNotExists(t *testing.T) {
	list := []string{"a", "b"}
	list = AddIfNotExists("c", list)
	if len(list) != 3 || list[2] != "c" {
		t.Errorf("AddIfNotExists failed to append new element: %v", list)
	}
	list = AddIfNotExists("b", list)
	if len(list) != 3 {
		t.Errorf("AddIfNotExists modified slice when element already existed: %v", list)
	}
}

func TestAddIfNotExistsGeneric(t *testing.T) {
	list := []interface{}{1, "two"}
	list = AddIfNotExistsGeneric(3, list)
	if len(list) != 3 || list[2] != 3 {
		t.Errorf("AddIfNotExistsGeneric failed to append new element: %v", list)
	}
	list = AddIfNotExistsGeneric("two", list)
	if len(list) != 3 {
		t.Errorf("AddIfNotExistsGeneric modified slice when element already existed: %v", list)
	}
}

func TestStringInSlice(t *testing.T) {
	list := []string{"a", "b"}
	if !StringInSlice("a", list) {
		t.Error("StringInSlice did not find existing element")
	}
	if StringInSlice("c", list) {
		t.Error("StringInSlice found non-existent element")
	}
}
