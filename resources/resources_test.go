package resources

import (
	"reflect"
	"testing"
)

func Test_newResourceBox(t *testing.T) {
	t.Parallel()

	got, ok := NewResourceBox().(Box)

	if !ok {
		t.Errorf("NewResourceBox() = %v, want %v", got, "Box")
	}
}

func Test_ResourcesVariable(t *testing.T) {
	t.Parallel()

	got, ok := Resources.(Box)

	if !ok {
		t.Errorf("Resources = %v, want %v", got, "Box")
	}
}

func Test_resourcesBox_Get(t *testing.T) {
	t.Parallel()

	resources := NewResourceBox()

	content, ok := resources.Get("foo")

	if ok {
		t.Error("Non-existing element must returns FALSE in 2nd value")
	}

	if content != nil {
		t.Error("Non-existing element must returns NIL in 1st value")
	}

	value := []byte{1, 2}
	resources.Add("foo", value)

	content, ok = resources.Get("foo")

	if !ok {
		t.Error("Existing element must returns TRUE in 2nd value")
	}

	if !reflect.DeepEqual(content, value) {
		t.Errorf("Expecded value for existinf element is %v, got %v", value, content)
	}
}

func Test_resourcesBox_Has(t *testing.T) {
	t.Parallel()

	resources := NewResourceBox()

	if resources.Has("foo") {
		t.Error("Non-existing element must returns FALSE")
	}

	resources.Add("foo", []byte{})

	if !resources.Has("foo") {
		t.Error("Existing element must returns TRUE")
	}
}
