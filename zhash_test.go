package zhash

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
)

var testMap = map[string]interface{}{
	"float":          10.1,
	"float64":        10.2,
	"int":            10,
	"int64":          int64(11),
	"bool":           true,
	"bool_f":         false,
	"string":         "some text",
	"intSlice":       []int64{20, 22, 24},
	"interfaceSlice": []interface{}{20, 22, 24},
	"strSlice":       []string{"a", "b", "c"},
	"strISlice":      []interface{}{"a", "b", "c"},
	"mixedSlice":     []interface{}{"a", 1, "c"},
	"map": map[string]interface{}{
		"val1": 10,
		"val2": 11.0,
	},
}

func TestReadHashSuccess(t *testing.T) {
	fd, err := os.Open("test.json")
	if err != nil {
		t.Error("Cannot read test.json!")
	}
	defer fd.Close()

	h := NewHash(nil, json.Unmarshal)
	err = h.ReadHash(fd)
	if err != nil {
		t.Error("Error reading hash!")
	}
}

func TestReadHashFailParse(t *testing.T) {
	fd, err := os.Open("test_corrupted.json")
	if err != nil {
		t.Error("Cannot read test_corrupted.json!")
	}
	defer fd.Close()

	h := NewHash(nil, json.Unmarshal)
	err = h.ReadHash(fd)
	if err == nil {
		t.Error("Hash readed from corrupted source!")
	}
}

func TestReadHashFailNoUnmarshaller(t *testing.T) {
	fd, err := os.Open("test_corrupted.json")
	if err != nil {
		t.Error("Cannot read test_corrupted.json!")
	}
	defer fd.Close()

	h := NewHash(nil, nil)
	err = h.ReadHash(fd)
	if err == nil {
		t.Error("Hash readed from corrupted source!")
	}
}

type corruptedReader string

func (c corruptedReader) Read(b []byte) (int, error) {
	return 0, errors.New("I'm corrupted!")
}

func TestReadHashFailReaderErr(t *testing.T) {
	r := corruptedReader("aaA")

	h := NewHash(nil, nil)
	h.SetUnmarshallerFunc(json.Unmarshal)
	err := h.ReadHash(r)
	if err == nil {
		t.Error("Hash readed from corrupted source!")
	}
}

var setTests = []struct {
	path  string
	value interface{}
}{
	{"string", "s@t.r"},
	{"map.val1", 10},
	{"map.val2", map[string]interface{}{"provider": "bar", "pool": "baz"}},
	{"foo.bar.baz", 10.1},
}

func TestSetPath(t *testing.T) {
	m := map[string]interface{}{}
	conf := HashFromMap(m, nil, nil)
	for _, test := range setTests {
		conf.SetPath(test.value, test.path)
	}

	for i, test := range setTests {
		val := conf.GetPath(strings.Split(test.path, ".")...)
		if !reflect.DeepEqual(val, test.value) {
			t.Errorf("#%d: conf[%s]%#v != %#v", i, test.path, val, test.value)
		}
	}
}

func TestHashReader(t *testing.T) {
	conf := HashFromMap(testMap, nil, nil)
	r := conf.Reader()
	f, err := os.OpenFile(os.DevNull, os.O_RDWR, 0666)
	if err != nil {
		t.Error("Error opening DevNull:", err)
	}
	defer f.Close()
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		t.Error("Error reading from conf.Reader():", err)
	}
	f.Write(buf)
}

type TestGet struct {
	path  []string
	value interface{}
}

func TestGetPath(t *testing.T) {
	var getTests []TestGet = []TestGet{
		{[]string{"string"}, "some text"},
		{[]string{"float"}, 10.1},
		{[]string{"map", "val1"}, 10},
		{[]string{"map", "val3"}, nil},
		{[]string{}, nil},
	}
	hash := HashFromMap(testMap, nil, nil)

	for i, test := range getTests {
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

func TestGetMapSuccess(t *testing.T) {
	conf := HashFromMap(testMap, nil, nil)
	var mapGetTests []TestGet = []TestGet{
		{[]string{"map"}, map[string]interface{}{
			"val1": 10,
			"val2": 11.0,
		}},
	}

	for i, test := range mapGetTests {
		m, err := conf.GetMap(test.path...)
		if err != nil {
			t.Errorf("#%d: GetMap(%s) caused error: %v", i, test.path, err)
		}
		if !reflect.DeepEqual(m, test.value) {
			t.Errorf("#%d: GetMap(%s)=%#v; want %#v", i, test.path, m, test.value)
		}
	}
}

func TestGetMapFail(t *testing.T) {
	conf := HashFromMap(testMap, nil, nil)
	var mapGetTests []TestGet = []TestGet{
		{[]string{"intSlice"}, map[string]interface{}{}},
		{[]string{"getter"}, map[string]interface{}{}},
	}

	for i, test := range mapGetTests {
		m, err := conf.GetMap(test.path...)
		if err == nil {
			t.Errorf("#%d: GetMap(%s) doesn't cause error, but it should", i, test.path)
		} else {
			t.Logf("Err: %s", err)
		}
		if !reflect.DeepEqual(m, test.value) {
			t.Errorf("#%d: GetMap(%s)=%#v; want %#v", i, test.path, m, test.value)
		}
	}
}

func TestGetSliceSuccess(t *testing.T) {
	conf := HashFromMap(testMap, nil, nil)
	var sliceGetTests []TestGet = []TestGet{
		{[]string{"interfaceSlice"}, []interface{}{20, 22, 24}},
	}

	for i, test := range sliceGetTests {
		m, err := conf.GetSlice(test.path...)
		if err != nil {
			t.Errorf("#%d: GetSlice(%s) caused error: %v", i, test.path, err)
		}
		if !reflect.DeepEqual(m, test.value) {
			t.Errorf("#%d: GetSlice(%s)=%#v; want %#v", i, test.path, m, test.value)
		}
	}
}

func TestGetSliceFail(t *testing.T) {
	conf := HashFromMap(testMap, nil, nil)
	var sliceGetTests []TestGet = []TestGet{
		{[]string{"int"}, []interface{}{}},
		{[]string{"intSlice"}, []interface{}{}},
		{[]string{"not", "exists"}, []interface{}{}},
	}

	for i, test := range sliceGetTests {
		m, err := conf.GetSlice(test.path...)
		if err == nil {
			t.Errorf("#%d: GetSlice(%s) doesn't cause error, but it should", i, test.path)
		} else {
			t.Logf("Err: %s", err)
		}
		if !reflect.DeepEqual(m, test.value) {
			t.Errorf("#%d: GetSlice(%s)=%#v; want %#v", i, test.path, m, test.value)
		}
	}
}

func TestGetStringSliceSuccess(t *testing.T) {
	conf := HashFromMap(testMap, nil, nil)
	var stringSliceGetTests []TestGet = []TestGet{
		{[]string{"strSlice"}, []string{"a", "b", "c"}},
		{[]string{"strISlice"}, []string{"a", "b", "c"}},
	}

	for i, test := range stringSliceGetTests {
		m, err := conf.GetStringSlice(test.path...)
		if err != nil {
			t.Errorf("#%d: GetStringSlice(%s) caused error: %v", i, test.path, err)
		}
		if !reflect.DeepEqual(m, test.value) {
			t.Errorf("#%d: GetStringSlice(%s)=%#v; want %#v", i, test.path, m, test.value)
		}
	}
}

func TestGetStringSliceFail(t *testing.T) {
	conf := HashFromMap(testMap, nil, nil)
	var stringSliceGetTests []TestGet = []TestGet{
		{[]string{"intSlice"}, []string{}},
		{[]string{"mixedSlice"}, []string{}},
		{[]string{"meta", "foo", "bar"}, []string{}},
		{[]string{"map"}, []string{}},
	}

	for i, test := range stringSliceGetTests {
		m, err := conf.GetStringSlice(test.path...)
		if err == nil {
			t.Errorf("#%d: GetStringSlice(%s) doesn't cause error, but it should", i, test.path)
		} else {
			t.Logf("Err: %s", err)
		}
		if !reflect.DeepEqual(m, test.value) {
			t.Errorf("#%d: GetStringSlice(%s)=%#v; want %#v", i, test.path, m, test.value)
		}
	}
}

func TestGetIntSuccess(t *testing.T) {
	var intGetTests []TestGet = []TestGet{
		{[]string{"int"}, int64(10)},
		{[]string{"map", "val1"}, int64(12)},
	}

	h := map[string]interface{}{
		"int": 10,
		"map": map[string]interface{}{"val1": int64(12)},
	}

	hash := HashFromMap(h, nil, nil)

	for i, test := range intGetTests {
		in, err := hash.GetInt(test.path...)
		if err != nil {
			v := hash.GetPath(test.path...)
			t.Errorf("#%d: GetInt(%s) caused error: %v, %s.(type) = %T", i, test.path, err, test.path, v)
		}
		if in != test.value {
			t.Errorf("#%d: GetInt(%s)=%d; want %d", i, test.path, in, test.value)
		}
	}
}

func TestGetIntFail(t *testing.T) {
	conf := HashFromMap(testMap, nil, nil)
	var intGetTests []TestGet = []TestGet{
		{[]string{"meta", "foo", "bar"}, int64(0)},
		{[]string{"domain"}, int64(0)},
	}

	for i, test := range intGetTests {
		in, err := conf.GetInt(test.path...)
		if err == nil {
			t.Errorf("#%d: GetInt(%s) doesn't cause error, but it should", i, test.path)
		} else {
			t.Logf("Err: %s", err)
		}
		if in != test.value {
			t.Errorf("#%d: GetInt(%s)=%d; want %d", i, test.path, in, test.value)
		}
	}
}

func TestGetFloatSuccess(t *testing.T) {
	conf := HashFromMap(testMap, nil, nil)
	var floatGetTests []TestGet = []TestGet{
		{[]string{"float"}, 10.1},
		{[]string{"int"}, 10.0},
		{[]string{"int64"}, 11.0},
	}

	for i, test := range floatGetTests {
		f, err := conf.GetFloat(test.path...)
		if f != test.value {
			t.Errorf("#%d: GetFloat(%s)=%f; want %v", i, test.path, f, test.value)
		}
		if err != nil {
			t.Errorf("#%d: GetFloat(%s) returned error: %s", i, test.path, err)
		}
	}
}

func TestGetFloatFail(t *testing.T) {
	conf := HashFromMap(testMap, nil, nil)
	var floatGetTests []TestGet = []TestGet{
		{[]string{"foo", "bar"}, 0.0},
		{[]string{"bool"}, 0.0},
	}

	for i, test := range floatGetTests {
		f, err := conf.GetFloat(test.path...)
		if f != test.value {
			t.Errorf("#%d: GetFloat(%s)=%f; want %v", i, test.path, f, test.value)
		}
		if err == nil {
			t.Errorf("#%d: GetFloat(%s) doesn't return any error, but it should fail", i, test.path)
		}
	}
}

func TestGetStringSuccess(t *testing.T) {
	conf := HashFromMap(testMap, nil, nil)
	var stringGetTests []TestGet = []TestGet{
		{[]string{"string"}, "some text"},
	}

	for i, test := range stringGetTests {
		f, err := conf.GetString(test.path...)
		if f != test.value {
			t.Errorf("#%d: GetString(%s)=%s; want %v", i, test.path, f, test.value)
		}
		if err != nil {
			t.Errorf("#%d: GetString(%s) returned error: %s", i, test.path, err)
		}
	}
}

func TestGetStringFail(t *testing.T) {
	conf := HashFromMap(testMap, nil, nil)
	var stringGetTests []TestGet = []TestGet{
		{[]string{"bar", "baz"}, ""},
		{[]string{"bool"}, ""},
		{[]string{"int"}, ""},
	}

	for i, test := range stringGetTests {
		f, err := conf.GetString(test.path...)
		if f != test.value {
			t.Errorf("#%d: GetString(%s)=%s; want %v", i, test.path, f, test.value)
		}
		if err == nil {
			t.Errorf("#%d: GetString(%s) doesn't return any error, but it should fail", i, test.path)
		}
	}
}

func TestGetBoolSuccess(t *testing.T) {
	conf := HashFromMap(testMap, nil, nil)
	var stringGetTests []TestGet = []TestGet{
		{[]string{"bool"}, true},
		{[]string{"bool_f"}, false},
	}

	for i, test := range stringGetTests {
		f, err := conf.GetBool(test.path...)
		if f != test.value {
			t.Errorf("#%d: GetBool(%s)=%s; want %v", i, test.path, f, test.value)
		}
		if err != nil {
			t.Errorf("#%d: GetBool(%s) returned error: %s", i, test.path, err)
		}
	}
}

func TestGetBoolFail(t *testing.T) {
	conf := HashFromMap(testMap, nil, nil)
	var stringGetTests []TestGet = []TestGet{
		{[]string{"int"}, false},
		{[]string{"map", "val3"}, false},
	}

	for i, test := range stringGetTests {
		b, err := conf.GetBool(test.path...)
		t.Logf("b=\"%#v\", b.(type)=\"%T\"", b, b)
		if b != test.value {
			t.Errorf("#%d: GetBool(%s)=%v; want %v", i, test.path, b, test.value)
		}
		if err == nil {
			t.Errorf("#%d: GetBool(%s) doesn't return any error, but it should fail", i, test.path)
		}
	}
}

func TestToStringSuccess(t *testing.T) {
	conf := HashFromMap(testMap, nil, nil)

	t.Logf("Hash: %s", conf)
}

type buggyStruct struct {
	Id int `json:"id"`
}

func (b buggyStruct) MarshalJSON() ([]byte, error) {
	return nil, errors.New("Baka!")
}

func TestToStringFail(t *testing.T) {
	conf := NewHash(nil, nil)

	value := buggyStruct{Id: 10}
	conf.Set(value, "meta", "bug")

	t.Logf("Hash: %s", conf)
}

func TestToJson(t *testing.T) {
	hash := HashFromMap(map[string]interface{}{
		"rec1": "val one",
		"rec2": map[string]interface{}{
			"sub_rec1": 2,
			"sub_rec2": "string",
		},
	}, json.Marshal, json.Unmarshal)

	jsonText := "{\"rec1\":\"val one\",\"rec2\":{\"sub_rec1\":2,\"sub_rec2\":\"string\"}}"

	convert, err := json.Marshal(hash)
	if err != nil {
		t.Error("Error marshalling hash to json:", err)
	}

	if string(convert) != jsonText {
		t.Errorf("Marshalled json differs from wanted:\nWant: %s\nGot: %s",
			jsonText, string(convert))
	}
}

func TestWriteHash(t *testing.T) {
	conf := HashFromMap(testMap, nil, nil)
	conf.SetMarshallerFunc(json.Marshal)
	f, _ := os.OpenFile(os.DevNull, os.O_RDWR, 0666)
	defer f.Close()
	if err := conf.WriteHash(f); err != nil {
		t.Errorf("Errors while writing config: %s", err.Error())
	}
}

func TestWriteHashError(t *testing.T) {
	conf := NewHash(json.Marshal, nil)
	conf.Set(buggyStruct{10}, "bug")
	f, _ := os.OpenFile(os.DevNull, os.O_RDWR, 0666)
	defer f.Close()
	if err := conf.WriteHash(f); err == nil {
		t.Errorf("No error while marshalling buggyStruct!")
	}
}
