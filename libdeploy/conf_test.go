package libdeploy

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
)

type recoverCallback func(r interface{})

func failOnRecover(callback recoverCallback) {
	if r := recover(); r != nil {
		callback(r)
	}
}

func mapsEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}

	if (a == nil && b != nil) || (b == nil && a != nil) {
		return false
	}

	result := true
	switch ma := a.(type) {
	case map[string]interface{}:
		switch mb := b.(type) {
		case map[string]interface{}:
			for na, va := range ma {
				switch val := va.(type) {
				case map[string]interface{}:
					result = mapsEqual(val, mb[na])
				default:
					result = (va == mb[na])
				}
				if !result {
					return result
				}
			}
		default:
			return false
		}
	default:
		return false
	}

	return true
}

func slicesEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}

	if (a == nil && b != nil) || (b == nil && a != nil) {
		return false
	}

	result := true
	switch sa := a.(type) {
	case []interface{}:
		switch sb := b.(type) {
		case []interface{}:
			if len(sa) != len(sb) {
				return false
			}
			for i, va := range sa {
				result = (va == sb[i])
				if !result {
					return result
				}
			}
		default:
			return false
		}
	default:
		return false
	}

	return true
}

var conf Config

func TestReadConfig(t *testing.T) {
	defer failOnRecover(func(r interface{}) { t.Errorf("ReadConfig caused panic: %v", r) })
	fd, err := os.Open("test.toml")
	if err != nil {
		t.Errorf("%#v", err)
	}
	defer fd.Close()

	err = conf.ReadConfig(fd)

	if err != nil {
		t.Error(err)
	}
}

func TestValidateError(t *testing.T) {
	defer failOnRecover(func(r interface{}) { t.Errorf("Validate caused panic") })
	errs := conf.Validate()
	if len(errs) == 0 {
		t.Error("Test validate error, there are reqired parameters in config")
	}

	for _, err := range errs {
		t.Log(err.Error())
	}
}

type ParseSetArgsTest struct {
	pathval string
	path    string
	val     interface{}
}

var parseSetArgsTests = []ParseSetArgsTest{
	{"setter.time:2014-05-09T12:01:05Z", "setter.time", time.Date(2014, 05, 9, 12, 01, 05, 0, time.UTC)},
	{"setter.int:214", "setter.int", 214},
	{"setter.float:21.4", "setter.float", 21.4},
	{"setter.bool:true", "setter.bool", true},
	{"setter.string:Tests env", "setter.string", "Tests env"},
}

func TestParseSetArgs(t *testing.T) {
	var i int
	var test ParseSetArgsTest
	defer failOnRecover(func(r interface{}) { t.Errorf("#%d: ParseSetArgument(%s) cause panic", i, test.pathval) })
	for i, test = range parseSetArgsTests {
		v, p := ParseSetArgument(test.pathval)
		if v != test.val || p != test.path {
			t.Errorf("#%d: ParseSetArgument(%s)=%#v,%#v; want %#v, %#v", i, test.pathval, v, p, test.val, test.path)
		}
	}
}

type SetTest struct {
	path  string
	value interface{}
}

var setTests []SetTest = []SetTest{
	{"meta.email", "s@t.r"},
	{"meta.bar", 10},
	{"resources.foo", map[string]interface{}{"provider": "bar", "pool": "baz"}},
	{"foo.bar.baz", 10.1},
}

func TestSetPath(t *testing.T) {
	var i int
	var test SetTest
	defer failOnRecover(func(r interface{}) { t.Errorf("#%d: SetPath(%s) cause panic", i, test.path) })
	for i, test = range setTests {
		conf.SetPath(test.value, test.path)
	}
}

func TestValidateSuccess(t *testing.T) {
	defer failOnRecover(func(r interface{}) { t.Errorf("Validate caused panic") })
	errs := conf.Validate()
	if len(errs) != 0 {
		t.Error("Test validate error, config is correct")
		for _, err := range errs {
			t.Log(err.Error())
		}
	}
}

func TestWriteConfig(t *testing.T) {
	defer failOnRecover(func(r interface{}) { t.Errorf("WriteConfig caused panic") })
	f, _ := os.OpenFile(os.DevNull, os.O_RDWR, 0666)
	defer f.Close()
	if err := conf.WriteConfig(f); err != nil {
		t.Errorf("Errors while writing config: %s", err.Error())
	}
}

func TestConfigReader(t *testing.T) {
	defer failOnRecover(func(r interface{}) { t.Errorf("Reader caused panic") })
	r := conf.Reader()
	f, _ := os.OpenFile(os.DevNull, os.O_RDWR, 0666)
	defer f.Close()
	buf, _ := ioutil.ReadAll(r)
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
	var i int
	var test TestGet
	defer failOnRecover(func(r interface{}) { t.Errorf("GetPath #%d: GetPaht(%s) caused panic: %v", r) })
	for i, test = range getTests {
		value := conf.GetPath(test.path...)
		if value != test.value {
			t.Errorf("#%d: GetPath(%s)=%#v; want %#v", i, test.path, value, test.value)
		}
	}
}

var mapGetTests []TestGet = []TestGet{
	{[]string{"meta"}, map[string]interface{}{
		"owner":       "e.persienko",
		"email":       "s@t.r",
		"description": "Tests",
		"bar":         10,
	}, false},
	{[]string{"meta", "foo", "bar"}, map[string]interface{}{}, false},
	{[]string{"domain"}, map[string]interface{}{}, true},
}

func TestGetMap(t *testing.T) {
	var i int
	var test TestGet
	for i, test = range mapGetTests {
		m, err := conf.GetMap(test.path...)
		if !mapsEqual(m, test.value) {
			t.Errorf("#%d: GetMap(%s)=%#v, %v; want %#v, %v", i, test.path, m, err, test.value, test.fails)
		}
		if test.fails {
			if err == nil {
				t.Errorf("#%d: GetMap(%s) doesn't return any error, but it should fail", i, test.path)
			}
		}
	}
}

var sliceGetTests []TestGet = []TestGet{
	{[]string{"resources", "conf", "depends"}, []interface{}{"mysql_single", "mongo_single", "node_single"}, false},
	{[]string{"meta", "foo", "bar"}, []interface{}{}, false},
	{[]string{"domain"}, []interface{}{}, true},
}

func TestGetSlice(t *testing.T) {
	var i int
	var test TestGet
	for i, test = range sliceGetTests {
		m, err := conf.GetSlice(test.path...)
		if !slicesEqual(m, test.value) {
			t.Errorf("#%d: GetSlice(%s)=%#v, %v; want %#v, %v", i, test.path, m, err, test.value, test.fails)
		}
		if test.fails {
			if err == nil {
				t.Errorf("#%d: GetSlice(%s) doesn't return any error, but it should fail", i, test.path)
			}
		}
	}
}

var intGetTests []TestGet = []TestGet{
	{[]string{"meta", "bar"}, 10, false},
	{[]string{"meta", "foo", "bar"}, 0, false},
	{[]string{"domain"}, 0, true},
}

func TestGetInt(t *testing.T) {
	var i int
	var test TestGet
	defer failOnRecover(func(r interface{}) { t.Errorf("GetInt #%d: GetInt(%s) caused panic: %v", i, test.path, r) })
	for i, test = range intGetTests {
		in, err := conf.GetInt(test.path...)
		if in != test.value {
			t.Errorf("#%d: GetInt(%s)=%d, %v; want %d, %v", i, test.path, in, err, test.value, test.fails)
		}
		if test.fails {
			if err == nil {
				t.Errorf("#%d: GetInt(%s) doesn't return any error, but it should fail", i, test.path)
			}
		}
	}
}

var floatGetTests []TestGet = []TestGet{
	{[]string{"foo", "bar", "baz"}, 10.1, false},
	{[]string{"meta", "bar"}, 10.0, false},
	{[]string{"meta", "foo", "bar"}, 0.0, false},
	{[]string{"domain"}, 0.0, true},
}

func TestGetFloat(t *testing.T) {
	var i int
	var test TestGet
	defer failOnRecover(func(r interface{}) { t.Errorf("GetFloat #%d: GetFloat(%s) caused panic: %v", i, test.path, r) })
	for i, test = range floatGetTests {
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
	{[]string{"meta", "bar", "bazzar"}, "", false},
	{[]string{"meta"}, "", true},
}

func TestGetString(t *testing.T) {
	var i int
	var test TestGet
	defer failOnRecover(func(r interface{}) { t.Errorf("#%d: GetString(%s) caused panic: %v", i, test.path, r) })
	for i, test = range stringGetTests {
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
