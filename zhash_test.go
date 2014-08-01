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

func readHash(path string) (Hash, error) {
	conf := NewHash()
	fd, err := os.Open(path)
	if err != nil {
		return conf, err
	}
	defer fd.Close()
	err = conf.ReadHash(fd)
	if err != nil {
		return NewHash(), err
	}
	return conf, nil
}

var configs = map[string]string{
	"valid":   "test_valid.toml",
	"invalid": "test_invalid.toml",
}

func TestValidateValid(t *testing.T) {
	conf, err := readHash(configs["valid"])
	if err != nil {
		t.Errorf("Error loading config: %v", err)
	}
	errs := conf.Validate()

	if len(errs) > 0 {
		t.Errorf("config doesn't validates, but it should; Errors: %v", errs)
	}
}

func TestValidateInvalid(t *testing.T) {
	conf, err := readHash(configs["invalid"])
	if err != nil {
		t.Errorf(" Error loading config: %v", err)
	}
	errs := conf.Validate()

	if len(errs) == 0 {
		t.Errorf("config validates, but it should fail")
	} else {
		for _, err = range errs {
			t.Log(err.Error())
		}
	}
}

var setTests = []struct {
	path  string
	value interface{}
}{
	{"meta.email", "s@t.r"},
	{"meta.bar", 10},
	{"resources.foo", map[string]interface{}{"provider": "bar", "pool": "baz"}},
	{"foo.bar.baz", 10.1},
}

func TestSetPath(t *testing.T) {
	conf, err := readHash(configs["valid"])
	if err != nil {
		t.Errorf("Error loading config: %v", err)
	}
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

func TestWriteHash(t *testing.T) {
	conf, err := readHash(configs["valid"])
	if err != nil {
		t.Errorf("Error loading config: %v", err)
	}
	f, _ := os.OpenFile(os.DevNull, os.O_RDWR, 0666)
	defer f.Close()
	if err := conf.WriteHash(f); err != nil {
		t.Errorf("Errors while writing config: %s", err.Error())
	}
}

func TestHashReader(t *testing.T) {
	conf, err := readHash(configs["valid"])
	if err != nil {
		t.Errorf("Error loading config: %v", err)
	}
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
		{[]string{"domain"}, "t6"},
		{[]string{"meta", "owner"}, "e.persienko"},
		{[]string{"resources", "mongo_single", "provider"}, "dbfarm"},
		{[]string{"meta", "foo", "bar"}, nil},
		{[]string{}, nil},
	}

	conf, err := readHash(configs["valid"])
	if err != nil {
		t.Errorf("Error loading config: %v", err)
	}

	for i, test := range getTests {
		value := conf.GetPath(test.path...)
		if value != test.value {
			t.Errorf("#%d: GetPath(%s)=%#v; want %#v", i, test.path, value, test.value)
		}
	}
}

func TestGetMapSuccess(t *testing.T) {
	var mapGetTests []TestGet = []TestGet{
		{[]string{"meta"}, map[string]interface{}{
			"owner":       "e.persienko",
			"email":       "some@test.ru",
			"description": "Tests",
			"bool":        true,
			"bool_f":      false,
		}},
	}

	conf, err := readHash(configs["valid"])
	if err != nil {
		t.Errorf("Error loading config: %v", err)
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
	var mapGetTests []TestGet = []TestGet{
		{[]string{"meta", "foo", "bar"}, map[string]interface{}{}},
		{[]string{"domain"}, map[string]interface{}{}},
	}
	conf, err := readHash(configs["valid"])
	if err != nil {
		t.Errorf("Error loading config: %v", err)
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
	var sliceGetTests []TestGet = []TestGet{
		{[]string{"resources", "conf", "depends"}, []interface{}{"mysql_single", "mongo_single", "node_single"}},
	}

	conf, err := readHash(configs["valid"])
	if err != nil {
		t.Errorf("Error loading config: %v", err)
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
	var sliceGetTests []TestGet = []TestGet{
		{[]string{"meta", "foo", "bar"}, []interface{}{}},
		{[]string{"domain"}, []interface{}{}},
	}

	conf, err := readHash(configs["valid"])
	if err != nil {
		t.Errorf("Error loading config: %v", err)
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
	var stringSliceGetTests []TestGet = []TestGet{
		{[]string{"resources", "conf", "depends"}, []string{"mysql_single", "mongo_single", "node_single"}},
	}
	conf, err := readHash(configs["valid"])
	if err != nil {
		t.Errorf("Error loading config: %v", err)
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
	var stringSliceGetTests []TestGet = []TestGet{
		{[]string{"getters", "intSlice"}, []string{}},
		{[]string{"meta", "foo", "bar"}, []string{}},
		{[]string{"domain"}, []string{}},
	}
	conf, err := readHash(configs["valid"])
	if err != nil {
		t.Errorf("Error loading config: %v", err)
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
		{[]string{"getters", "int"}, int64(10)},
	}
	conf, err := readHash(configs["valid"])
	if err != nil {
		t.Errorf("Error loading config: %v", err)
	}

	for i, test := range intGetTests {
		in, err := conf.GetInt(test.path...)
		if err != nil {
			v := conf.GetPath(test.path...)
			t.Errorf("#%d: GetInt(%s) caused error: %v, %s.(type) = %T", i, test.path, err, test.path, v)
		}
		if in != test.value {
			t.Errorf("#%d: GetInt(%s)=%d; want %d", i, test.path, in, test.value)
		}
	}
}

func TestGetIntFail(t *testing.T) {
	var intGetTests []TestGet = []TestGet{
		{[]string{"meta", "foo", "bar"}, int64(0)},
		{[]string{"domain"}, int64(0)},
	}
	conf, err := readHash(configs["valid"])
	if err != nil {
		t.Errorf("Error loading config: %v", err)
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
	var floatGetTests []TestGet = []TestGet{
		{[]string{"getters", "float"}, 10.1},
		{[]string{"getters", "int"}, 10.0},
	}

	conf, err := readHash(configs["valid"])
	if err != nil {
		t.Errorf("Error loading config: %v", err)
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
	var floatGetTests []TestGet = []TestGet{
		{[]string{"meta", "foo", "bar"}, 0.0},
		{[]string{"meta", "bool"}, 0.0},
	}

	conf, err := readHash(configs["valid"])
	if err != nil {
		t.Errorf("Error loading config: %v", err)
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
	var stringGetTests []TestGet = []TestGet{
		{[]string{"domain"}, "t6"},
	}
	conf, err := readHash(configs["valid"])
	if err != nil {
		t.Errorf("Error loading config: %v", err)
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
	var stringGetTests []TestGet = []TestGet{
		{[]string{"meta", "bar", "bazzar"}, ""},
		{[]string{"meta"}, ""},
	}
	conf, err := readHash(configs["valid"])
	if err != nil {
		t.Errorf("Error loading config: %v", err)
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
	var stringGetTests []TestGet = []TestGet{
		{[]string{"meta", "bool"}, true},
		{[]string{"meta", "bool_f"}, false},
	}
	conf, err := readHash(configs["valid"])
	if err != nil {
		t.Errorf("Error loading config: %v", err)
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
	var stringGetTests []TestGet = []TestGet{
		{[]string{"meta", "bar", "bazzar"}, false},
		{[]string{"meta"}, false},
	}
	conf, err := readHash(configs["valid"])
	if err != nil {
		t.Errorf("Error loading config: %v", err)
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
	conf, err := readHash(configs["valid"])
	if err != nil {
		t.Errorf("Error reading config")
	}

	t.Logf("Hash: %s", conf)
}

type buggyStruct struct {
	Id int `json:"id"`
}

func (b buggyStruct) MarshalJSON() ([]byte, error) {
	return nil, errors.New("Baka!")
}

func TestToStringFail(t *testing.T) {
	conf, err := readHash(configs["valid"])
	if err != nil {
		t.Errorf("Error reading config")
	}

	value := buggyStruct{Id: 10}
	conf.Set(value, "meta", "bug")

	t.Logf("Hash: %s", conf)
}

func TestMarshalToJSON(t *testing.T) {
	hash := HashFromMap(map[string]interface{}{
		"rec1": "val one",
		"rec2": map[string]interface{}{
			"sub_rec1": 2,
			"sub_rec2": "string",
		},
	})

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
