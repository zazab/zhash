package zhash

import (
	"reflect"
	"strings"
	"testing"
)

var testMap = map[string]interface{}{
	"float":        10.1,
	"float64":      10.2,
	"int":          10,
	"int64":        int64(11),
	"bool":         true,
	"bool_f":       false,
	"string":       "some text",
	"intSlice":     []int{10, 12, 14},
	"int64Slice":   []int64{20, 22, 24},
	"intISlice":    []interface{}{30, 32, 34},
	"intIMixSlice": []interface{}{30, int64(32), 34},
	"fltSlice":     []float64{40.1, 42.2, 44.3},
	"fltISlice":    []interface{}{50.1, 52.2, 54.3},
	"strSlice":     []string{"a", "b", "c"},
	"strISlice":    []interface{}{"a", "b", "c"},
	"mixedSlice":   []interface{}{"a", 1, "c"},
	"map": map[string]interface{}{
		"val1": 10,
		"val2": 11.0,
	},
	"toDel": 1,
	"map2": map[string]interface{}{
		"toDel": "a",
	},
}

func TestSetPath(t *testing.T) {
	tests := []struct {
		path  string
		value interface{}
	}{
		{"string", "s@t.r"},
		{"map.val1", 10},
		{"map.val2", map[string]interface{}{"provider": "bar", "pool": "baz"}},
		{"foo.bar.baz", 10.1},
	}

	m := map[string]interface{}{}
	hash := HashFromMap(m, nil, nil)
	for _, test := range tests {
		hash.SetPath(test.value, test.path)
	}

	for i, test := range tests {
		val := hash.GetPath(strings.Split(test.path, ".")...)
		if !reflect.DeepEqual(val, test.value) {
			t.Errorf("#%d: hash[%s]%#v != %#v", i, test.path, val, test.value)
		}
	}
}

func TestDelete(t *testing.T) {
	hash := HashFromMap(testMap, nil, nil)
	err := hash.Delete("toDel")
	if err != nil {
		t.Errorf("Error deleting toDel")
	}
	err = hash.Delete("map2", "toDel")
	if err != nil {
		t.Errorf("Error deleting map.toDel")
	}
	err = hash.Delete("map3", "toDel")
	if !IsNotFound(err) {
		t.Errorf("Delete from absent parent returned not notFoundError!")
	}

	err = hash.Delete("int", "toDel")
	if err == nil || IsNotFound(err) {
		t.Errorf("Delete from int parent returned wrong error (or no error)!")
	}
}

type getTest struct {
	path  []string
	value interface{}
	fails bool
}

func TestGetPath(t *testing.T) {
	tests := []getTest{
		{[]string{"string"}, "some text", false},
		{[]string{"float"}, 10.1, false},
		{[]string{"map", "val1"}, 10, false},
		{[]string{"map", "val3"}, nil, false},
		{[]string{}, nil, false},
	}
	hash := HashFromMap(testMap, nil, nil)

	for i, test := range tests {
		value := hash.GetPath(test.path...)
		if value != test.value {
			t.Errorf("#%d: GetPath(%s)=%#v; want %#v", i, test.path, value, test.value)
		}
	}
}

func TestNotFound(t *testing.T) {
	hash := HashFromMap(map[string]interface{}{"value": 10.1}, nil, nil)
	_, err := hash.GetInt("val")
	if !IsNotFound(err) {
		t.Errorf("IsNotFound returned false, but err is notFoundError")
	}
	_, err = hash.GetInt("value")
	if IsNotFound(err) {
		t.Errorf("IsNotFound returned true, but err is not notFoundError")
	}

}

func checkGet(n int, test getTest, v interface{}, err error, fn string, t *testing.T) {
	switch test.fails {
	case false:
		if err != nil {
			t.Errorf("#%d: %s(%s) caused error: %v", n, fn, test.path, err)
		}
	case true:
		if err == nil {
			t.Errorf("#%d: %s(%s) doesn't cause error, but it should", n, fn,
				test.path)
		} else {
			t.Logf("Err: %s", err)
		}
	}

	if !reflect.DeepEqual(v, test.value) {
		t.Errorf("#%d: %s(%s)=%#v; want %#v", n, fn, test.path, v, test.value)
	}
}

func TestGetMap(t *testing.T) {
	hash := HashFromMap(testMap, nil, nil)
	tests := []getTest{
		{[]string{"map"}, map[string]interface{}{
			"val1": 10,
			"val2": 11.0,
		}, false},
		{[]string{"intSlice"}, map[string]interface{}{}, true},
		{[]string{"getter"}, map[string]interface{}{}, true},
	}

	for i, test := range tests {
		m, err := hash.GetMap(test.path...)
		checkGet(i, test, m, err, "GetMap", t)
	}
}

func TestGetInt(t *testing.T) {
	tests := []getTest{
		{[]string{"int"}, int64(10), false},
		{[]string{"map", "val1"}, int64(12), false},
		{[]string{"meta", "foo", "bar"}, int64(0), true},
		{[]string{"domain"}, int64(0), true},
	}

	h := map[string]interface{}{
		"int": 10,
		"map": map[string]interface{}{"val1": int64(12)},
	}

	hash := HashFromMap(h, nil, nil)

	for i, test := range tests {
		in, err := hash.GetInt(test.path...)
		checkGet(i, test, in, err, "GetInt", t)
	}
}

func TestGetFloat(t *testing.T) {
	hash := HashFromMap(testMap, nil, nil)
	tests := []getTest{
		{[]string{"float"}, 10.1, false},
		{[]string{"int"}, 10.0, false},
		{[]string{"int64"}, 11.0, false},
		{[]string{"foo", "bar"}, 0.0, true},
		{[]string{"bool"}, 0.0, true},
	}

	for i, test := range tests {
		f, err := hash.GetFloat(test.path...)
		checkGet(i, test, f, err, "GetFloat", t)
	}
}

func TestGetString(t *testing.T) {
	hash := HashFromMap(testMap, nil, nil)
	tests := []getTest{
		{[]string{"string"}, "some text", false},
		{[]string{"bar", "baz"}, "", true},
		{[]string{"bool"}, "", true},
		{[]string{"int"}, "", true},
	}

	for i, test := range tests {
		s, err := hash.GetString(test.path...)
		checkGet(i, test, s, err, "GetString", t)
	}
}

func TestGetBool(t *testing.T) {
	hash := HashFromMap(testMap, nil, nil)
	tests := []getTest{
		{[]string{"bool"}, true, false},
		{[]string{"bool_f"}, false, false},
		{[]string{"int"}, false, true},
		{[]string{"map", "val3"}, false, true},
	}

	for i, test := range tests {
		b, err := hash.GetBool(test.path...)
		checkGet(i, test, b, err, "GetBool", t)
	}
}
