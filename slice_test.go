package zhash

import (
	"gopkg.in/yaml.v2"
	"reflect"
	"testing"
)

func copyTestMap() map[string]interface{} {
	cp := map[string]interface{}{}
	for key, val := range testMap {
		cp[key] = val
	}

	return cp
}

func TestGetSlice(t *testing.T) {
	hash := HashFromMap(testMap)
	tests := []getTest{
		{[]string{"strISlice"}, []interface{}{"a", "b", "c"}, false},
		{[]string{"int"}, []interface{}{}, true},
		{[]string{"intSlice"}, []interface{}{}, true},
		{[]string{"not", "exists"}, []interface{}{}, true},
	}

	for i, test := range tests {
		s, err := hash.GetSlice(test.path...)
		checkGet(i, test, s, err, "GetSlice", t)
	}
}

func TestGetIntSlice(t *testing.T) {
	hash := HashFromMap(testMap)
	tests := []getTest{
		{[]string{"intSlice"}, []int64{10, 12, 14}, false},
		{[]string{"int64Slice"}, []int64{20, 22, 24}, false},
		{[]string{"intISlice"}, []int64{30, 32, 34}, false},
		{[]string{"strISlice"}, []int64{}, true},
		{[]string{"mixedSlice"}, []int64{}, true},
		{[]string{"meta", "foo", "bar"}, []int64{}, true},
		{[]string{"map"}, []int64{}, true},
	}

	for i, test := range tests {
		s, err := hash.GetIntSlice(test.path...)
		checkGet(i, test, s, err, "GetIntSlice", t)
	}
}

func TestGetFloatSlice(t *testing.T) {
	hash := HashFromMap(testMap)
	tests := []getTest{
		{[]string{"fltSlice"}, []float64{40.1, 42.2, 44.3}, false},
		{[]string{"fltISlice"}, []float64{50.1, 52.2, 54.3}, false},
		{[]string{"intSlice"}, []float64{}, true},
		{[]string{"mixedSlice"}, []float64{}, true},
		{[]string{"meta", "foo", "bar"}, []float64{}, true},
		{[]string{"map"}, []float64{}, true},
	}

	for i, test := range tests {
		s, err := hash.GetFloatSlice(test.path...)
		checkGet(i, test, s, err, "GetFloatSlice", t)
	}
}

func TestGetStringSlice(t *testing.T) {
	hash := HashFromMap(testMap)
	tests := []getTest{
		{[]string{"strSlice"}, []string{"a", "b", "c"}, false},
		{[]string{"strISlice"}, []string{"a", "b", "c"}, false},
		{[]string{"intSlice"}, []string{}, true},
		{[]string{"mixedSlice"}, []string{}, true},
		{[]string{"meta", "foo", "bar"}, []string{}, true},
		{[]string{"map"}, []string{}, true},
	}

	for i, test := range tests {
		s, err := hash.GetStringSlice(test.path...)
		checkGet(i, test, s, err, "GetStringSlice", t)
	}
}

func TestGetMapSlice(t *testing.T) {
	testYaml := `
deploy:
        close: False
        framework: 224
        mysql: False
        mongo: True
        chmod:
          - path: config/ssh/id_rsa
            mode: 0600
          - path: config/ssh/ngs_rsa
            mode: 0600
          - path: config/ssl_certs_dir
            mode: 0600
            dirmode: 0700
            recursive: True
        chown:
          - path: config/ssh/ngs_rsa
            user: www-data
            group: www-logs
          - path: config/secret_dir_for_www-data
            user: www-data
            recursive: True
`
	rawMap := make(map[string]interface{})

	err := yaml.Unmarshal([]byte(testYaml), &rawMap)
	if err != nil {
		t.Fatalf("Error yaml unmarshall: %s", err.Error())
	}

	hash := HashFromMap(rawMap)

	result, err := hash.GetMapSlice("deploy", "chmod")
	if err != nil {
		t.Fatalf("Error getting map slice: %s", err.Error())
	}

	for i := 0; i < 3; i++ {
		t.Log(result[i]["path"])
	}
}

type appendSliceTest struct {
	set interface{}
	getTest
}

func checkAppend(n int, test appendSliceTest, s interface{}, err error, fn string, t *testing.T) {
	if test.fails {
		if err == nil {
			t.Errorf("#%d: %s doesn't fail, but should!", n, fn)
		}
	} else {
		if err != nil {
			t.Errorf("#%d: %s fails: %s", n, fn, err)
		}
	}

	if !reflect.DeepEqual(s, test.value) {
		t.Errorf("#%d: after %s hash.%s=%#v; want %#v", n, fn, test.path, s, test.value)
	}
}

func TestAppendSlice(t *testing.T) {
	tests := []appendSliceTest{
		{"d", getTest{[]string{"strISlice"}, []interface{}{"a", "b", "c", "d"}, false}},
		{36, getTest{[]string{"intISlice"}, []interface{}{30, 32, 34, 36}, false}},
		{36, getTest{[]string{"intIMixSlice"}, []interface{}{30, int64(32), 34, 36}, false}},
		{56.4, getTest{[]string{"fltISlice"}, []interface{}{50.1, 52.2, 54.3, 56.4}, false}},
		{888, getTest{[]string{"newSlice"}, []interface{}{888}, false}},
		{1, getTest{[]string{"fltSlice"}, []interface{}{}, true}},
	}

	hash := HashFromMap(copyTestMap())

	for i, test := range tests {
		var err error
		aerr := hash.AppendSlice(test.set, test.path...)
		s := []interface{}{}
		if !test.fails {
			s, err = hash.GetSlice(test.path...)
			if err != nil {
				t.Errorf("#%d: AppendSlice broke it! GetSlice fails with: %s",
					i, err)
			}
		}
		checkAppend(i, test, s, aerr, "AppendSlice", t)
	}
}

func TestAppendIntSlice(t *testing.T) {
	tests := []appendSliceTest{
		{int64(16), getTest{[]string{"intSlice"}, []int64{10, 12, 14, 16}, false}},
		{int64(26), getTest{[]string{"int64Slice"}, []int64{20, 22, 24, 26}, false}},
		{int64(36), getTest{[]string{"intISlice"}, []int64{30, 32, 34, 36}, false}},
		{int64(36), getTest{[]string{"intIMixSlice"}, []int64{30, 32, 34, 36}, false}},
		{int64(99), getTest{[]string{"newSlice"}, []int64{99}, false}},
		{int64(1), getTest{[]string{"fltSlice"}, []int64{}, true}},
	}

	hash := HashFromMap(copyTestMap())

	for i, test := range tests {
		var err error
		aerr := hash.AppendIntSlice(test.set.(int64), test.path...)
		s := []int64{}
		if !test.fails {
			s, err = hash.GetIntSlice(test.path...)
			if err != nil {
				t.Errorf("#%d: AppendIntSlice broke it! GetIntSlice fails with: %s",
					i, err)
			}
		}
		checkAppend(i, test, s, aerr, "AppendIntSlice", t)
	}
}

func TestAppendFloatSlice(t *testing.T) {
	tests := []appendSliceTest{
		{46.4, getTest{[]string{"fltSlice"}, []float64{40.1, 42.2, 44.3, 46.4}, false}},
		{56.4, getTest{[]string{"fltISlice"}, []float64{50.1, 52.2, 54.3, 56.4}, false}},
		{6.8, getTest{[]string{"newSlice"}, []float64{6.8}, false}},
		{1.5, getTest{[]string{"intSlice"}, []float64{}, true}},
	}

	hash := HashFromMap(copyTestMap())

	for i, test := range tests {
		var err error
		aerr := hash.AppendFloatSlice(test.set.(float64), test.path...)
		s := []float64{}
		if !test.fails {
			s, err = hash.GetFloatSlice(test.path...)
			if err != nil {
				t.Errorf("#%d: AppendFloatSlice broke it! GetSlice fails with: %s",
					i, err)
			}
		}
		checkAppend(i, test, s, aerr, "AppendFloatSlice", t)
	}
}

func TestAppendStringSlice(t *testing.T) {
	tests := []appendSliceTest{
		{"d", getTest{[]string{"strSlice"}, []string{"a", "b", "c", "d"}, false}},
		{"e", getTest{[]string{"strISlice"}, []string{"a", "b", "c", "e"}, false}},
		{"g", getTest{[]string{"newSlice"}, []string{"g"}, false}},
		{"m", getTest{[]string{"fltSlice"}, []string{}, true}},
	}

	hash := HashFromMap(copyTestMap())

	for i, test := range tests {
		var err error
		aerr := hash.AppendStringSlice(test.set.(string), test.path...)
		s := []string{}
		if !test.fails {
			s, err = hash.GetStringSlice(test.path...)
			if err != nil {
				t.Errorf("#%d: AppendStringSlice broke it! GetSlice fails with: %s",
					i, err)
			}
		}
		checkAppend(i, test, s, aerr, "AppendStringSlice", t)
	}
}
