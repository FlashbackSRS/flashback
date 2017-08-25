package model

import (
	"testing"

	"github.com/flimzy/diff"
)

type mockMC struct {
	ModelController
	t string
}

func (mc *mockMC) Type() string {
	return mc.t
}

const testMType = "foo"

var testMC = &mockMC{t: testMType}

func TestRegisterModelController(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		RegisterModelController(testMC)
		if _, ok := modelControllers[testMType]; !ok {
			t.Errorf("modelControllers not updated")
		}
		for _, t := range modelControllerTypes {
			if t == testMType {
				return
			}
		}
		t.Errorf("modelControllerTypes not updated")
	})
	t.Run("Duplicate", func(t *testing.T) {
		r := func() (r interface{}) {
			defer func() {
				r = recover()
			}()
			RegisterModelController(testMC)
			return
		}()
		expected := "A controller for 'foo' is already registered"
		if d := diff.Interface(expected, r); d != nil {
			t.Error(d)
		}
	})
}

func TestRegisteredModelControllers(t *testing.T) {
	expected := []string{"basic", "funcmapper", "foo"}
	result := RegisteredModelControllers()
	if d := diff.Interface(expected, result); d != nil {
		t.Error(d)
	}
}

func TestGetModelController(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mc, err := GetModelController(testMType)
		if err != nil {
			t.Fatal(err)
		}
		if mc != testMC {
			t.Errorf("Unexpected controller returned")
		}
	})
	t.Run("Not found", func(t *testing.T) {
		_, err := GetModelController("not found")
		if err == nil || err.Error() != "ModelController for 'not found' not found" {
			t.Errorf("Unexpected error: %s", err)
		}
	})
}
