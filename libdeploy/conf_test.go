package libdeploy

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"
)

type recoverCallback func(r interface{})

func failOnRecover(callback recoverCallback) {
	if r := recover(); r != nil {
		callback(r)
	}
}

func readConfig(path string) (Config, error) {
	conf := NewConfig()
	fd, err := os.Open(path)
	if err != nil {
		return conf, err
	}
	defer fd.Close()
	err = conf.ReadConfig(fd)
	if err != nil {
		return NewConfig(), err
	}
	return conf, nil
}

var configs = map[string]struct {
	path  string
	valid bool
}{
	"valid":   {"test_valid.toml", true},
	"invalid": {"test_invalid.toml", false},
}

func TestValidate(t *testing.T) {
	for c, test := range configs {
		conf, err := readConfig(test.path)
		if err != nil {
			t.Errorf("%s: Error loading config: %v", c, err)
		}
		errs := conf.Validate()

		if test.valid {
			if len(errs) > 0 {
				t.Errorf("%s: config doesn't validates, but it should; Errors: %v", c, errs)
			}
		} else {
			if len(errs) == 0 {
				t.Errorf("%s: config validates, but it should fail", c)
			} else {
				for _, err = range errs {
					t.Log(err.Error())
				}
			}
		}
	}
}

var parseSetArgsTests = []struct {
	pathval, path string
	val           interface{}
}{
	{"setter.time:2014-05-09T12:01:05Z", "setter.time", time.Date(2014, 05, 9, 12, 01, 05, 0, time.UTC)},
	{"setter.int:214", "setter.int", 214},
	{"setter.float:21.4", "setter.float", 21.4},
	{"setter.bool:true", "setter.bool", true},
	{"setter.string:Tests env", "setter.string", "Tests env"},
}

func TestParseSetArgs(t *testing.T) {
	for i, test := range parseSetArgsTests {
		v, p := ParseSetArgument(test.pathval)
		if v != test.val || p != test.path {
			t.Errorf("#%d: ParseSetArgument(%s)=%#v,%#v; want %#v, %#v", i, test.pathval, v, p, test.val, test.path)
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
	conf, err := readConfig(configs["valid"].path)
	if err != nil {
		t.Errorf("Error loading config: %v", err)
	}
	for _, test := range setTests {
		conf.SetPath(test.value, test.path)
	}
}

func TestWriteConfig(t *testing.T) {
	conf, err := readConfig(configs["valid"].path)
	if err != nil {
		t.Errorf("Error loading config: %v", err)
	}
	f, _ := os.OpenFile(os.DevNull, os.O_RDWR, 0666)
	defer f.Close()
	if err := conf.WriteConfig(f); err != nil {
		t.Errorf("Errors while writing config: %s", err.Error())
	}
}

func TestConfigReader(t *testing.T) {
	conf, err := readConfig(configs["valid"].path)
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
	fails bool
}

var getTests []TestGet = []TestGet{
	{[]string{"domain"}, "t6", false},
	{[]string{"meta", "owner"}, "e.persienko", false},
	{[]string{"resources", "mongo_single", "provider"}, "dbfarm", false},
	{[]string{"meta", "foo", "bar"}, nil, false},
}

func TestGetPath(t *testing.T) {
	conf, err := readConfig(configs["valid"].path)
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

var mapGetTests []TestGet = []TestGet{
	{[]string{"meta"}, map[string]interface{}{
		"owner":       "e.persienko",
		"email":       "some@test.ru",
		"description": "Tests",
	}, false},
	{[]string{"meta", "foo", "bar"}, map[string]interface{}{}, true},
	{[]string{"domain"}, map[string]interface{}{}, true},
}

func TestGetMap(t *testing.T) {
	conf, err := readConfig(configs["valid"].path)
	if err != nil {
		t.Errorf("Error loading config: %v", err)
	}

	for i, test := range mapGetTests {
		m, err := conf.GetMap(test.path...)
		if !test.fails {
			if err != nil {
				t.Errorf("#%d: GetMap(%s) caused error: %v", i, test.path, err)
			}
		} else {
			if err == nil {
				t.Errorf("#%d: GetMap(%s) doesn't cause error, but it should", i, test.path)
			}
		}
		if !reflect.DeepEqual(m, test.value) {
			t.Errorf("#%d: GetMap(%s)=%#v, %v; want %#v, %v", i, test.path, m, err, test.value, test.fails)
		}
	}
}

var sliceGetTests []TestGet = []TestGet{
	{[]string{"resources", "conf", "depends"}, []interface{}{"mysql_single", "mongo_single", "node_single"}, false},
	{[]string{"meta", "foo", "bar"}, []interface{}{}, true},
	{[]string{"domain"}, []interface{}{}, true},
}

func TestGetSlice(t *testing.T) {
	conf, err := readConfig(configs["valid"].path)
	if err != nil {
		t.Errorf("Error loading config: %v", err)
	}

	for i, test := range sliceGetTests {
		m, err := conf.GetSlice(test.path...)
		if !test.fails {
			if err != nil {
				t.Errorf("#%d: GetSlice(%s) caused error: %v", i, test.path, err)
			}
		} else {
			if err == nil {
				t.Errorf("#%d: GetSlice(%s) doesn't cause error, but it should", i, test.path)
			}
		}
		if !reflect.DeepEqual(m, test.value) {
			t.Errorf("#%d: GetSlice(%s)=%#v, %v; want %#v, %v", i, test.path, m, err, test.value, test.fails)
		}
	}
}

var stringSliceGetTests []TestGet = []TestGet{
	{[]string{"resources", "conf", "depends"}, []string{"mysql_single", "mongo_single", "node_single"}, false},
	{[]string{"getters", "intSlice"}, []string{}, true},
	{[]string{"meta", "foo", "bar"}, []string{}, true},
	{[]string{"domain"}, []string{}, true},
}

func TestGetStringSlice(t *testing.T) {
	conf, err := readConfig(configs["valid"].path)
	if err != nil {
		t.Errorf("Error loading config: %v", err)
	}

	for i, test := range stringSliceGetTests {
		m, err := conf.GetStringSlice(test.path...)
		if !test.fails {
			if err != nil {
				t.Errorf("#%d: GetStringSlice(%s) caused error: %v", i, test.path, err)
			}
		} else {
			if err == nil {
				t.Errorf("#%d: GetStringSlice(%s) doesn't cause error, but it should", i, test.path)
			}
		}
		if !reflect.DeepEqual(m, test.value) {
			t.Errorf("#%d: GetStringSlice(%s)=%#v, %v; want %#v, %v", i, test.path, m, err, test.value, test.fails)
		}
	}
}

var intGetTests []TestGet = []TestGet{
	{[]string{"getters", "int"}, int64(10), false},
	{[]string{"meta", "foo", "bar"}, int64(0), true},
	{[]string{"domain"}, int64(0), true},
}

func TestGetInt(t *testing.T) {
	conf, err := readConfig(configs["valid"].path)
	if err != nil {
		t.Errorf("Error loading config: %v", err)
	}

	for i, test := range intGetTests {
		in, err := conf.GetInt(test.path...)
		if !test.fails {
			if err != nil {
				v := conf.GetPath(test.path...)
				t.Errorf("#%d: GetInt(%s) caused error: %v, %s.(type) = %T", i, test.path, err, test.path, v)
			}
		} else {
			if err == nil {
				t.Errorf("#%d: GetInt(%s) doesn't cause error, but it should", i, test.path)
			}
		}
		if in != test.value {
			t.Errorf("#%d: GetInt(%s)=%d, %v; want %d, %v", i, test.path, in, err, test.value, test.fails)
		}
	}
}

var floatGetTests []TestGet = []TestGet{
	{[]string{"getters", "float"}, 10.1, false},
	{[]string{"getters", "int"}, 10.0, false},
	{[]string{"meta", "foo", "bar"}, 0.0, true},
	{[]string{"domain"}, 0.0, true},
}

func TestGetFloat(t *testing.T) {
	conf, err := readConfig(configs["valid"].path)
	if err != nil {
		t.Errorf("Error loading config: %v", err)
	}

	for i, test := range floatGetTests {
		f, err := conf.GetFloat(test.path...)
		if f != test.value {
			t.Errorf("#%d: GetFloat(%s)=%f, %v; want %v, %v", i, test.path, f, err, test.value, test.fails)
		}
		if test.fails {
			if err == nil {
				t.Errorf("#%d: GetFloat(%s) doesn't return any error, but it should fail", i, test.path)
			}
		}
	}
}

var stringGetTests []TestGet = []TestGet{
	{[]string{"domain"}, "t6", false},
	{[]string{"meta", "bar", "bazzar"}, "", true},
	{[]string{"meta"}, "", true},
}

func TestGetString(t *testing.T) {
	conf, err := readConfig(configs["valid"].path)
	if err != nil {
		t.Errorf("Error loading config: %v", err)
	}

	for i, test := range stringGetTests {
		f, err := conf.GetString(test.path...)
		if f != test.value {
			t.Errorf("#%d: GetString(%s)=%s, %v; want %v, %v", i, test.path, f, err, test.value, test.fails)
		}
		if test.fails {
			if err == nil {
				t.Errorf("#%d: GetString(%s) doesn't return any error, but it should fail", i, test.path)
			}
		}
	}
}
