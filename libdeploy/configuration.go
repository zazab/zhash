package configuration

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

var timeFormat = "2006-01-02T15:04:05Z"

func ReadConfig(fileName string) (config map[string]interface{}, err error) {
	blob, err := ioutil.ReadFile(fileName)
	if err != nil {
		return
	}

	_, err = toml.Decode(string(blob), &config)

	return
}

func SPrintConfig(config map[string]interface{}) (conf string, err error) {
	buf := new(bytes.Buffer)
	if err = toml.NewEncoder(buf).Encode(config); err != nil {
		return
	}

	conf = buf.String()
	return
}

func PrintConfig(config map[string]interface{}) (err error) {
	buf := new(bytes.Buffer)
	if err = toml.NewEncoder(buf).Encode(config); err != nil {
		return
	}

	fmt.Println(buf.String())
	return
}

func SetVariable(path string, value interface{}, config map[string]interface{}) {
	var buffer, changer map[string]interface{}
	var last_path string

	variable_path := strings.Split(path, ".")
	for num, p := range variable_path {
		if buffer == nil {
			if num+1 < len(variable_path) { // first element
				if config[p] == nil { // if no middle element
					config[p] = map[string]interface{}{}
				}
				buffer = config[p].(map[string]interface{})
			} else { // first and last
				changer = config
				last_path = p
			}
		} else {
			if num+1 < len(variable_path) { // middle element
				if buffer[p] == nil { // if no middle element
					buffer[p] = map[string]interface{}{}
				}
				buffer = buffer[p].(map[string]interface{})
			} else { // last element
				changer = buffer
				last_path = p
			}
		}
	}

	changer[last_path] = value
}

func ReplaceConfigParameter(path string, config map[string]interface{}) {
	buf := strings.SplitN(path, ":", 2)
	path = buf[0]
	val := buf[1]

	if t, err := time.Parse(timeFormat, val); err != nil {
		if i, err := strconv.Atoi(val); err != nil {
			if r, err := strconv.ParseFloat(val, 64); err != nil {
				if b, err := strconv.ParseBool(val); err != nil {
					SetVariable(path, val, config) // Cannot conver to any type, sujesting string
				} else { // Converted to bool
					SetVariable(path, b, config)
				}
			} else { // Converted to float
				SetVariable(path, r, config)
			}
		} else { // Converted to int
			SetVariable(path, i, config)
		}
	} else { // Converted to time
		SetVariable(path, t, config)
	}
}

func CheckRequired(conf interface{}, fullPath []string) (errs []error) {
	switch conf.(type) {
	case map[string]interface{}:
		for p, val := range conf.(map[string]interface{}) {
			errs = append(errs, CheckRequired(val, append(fullPath, p))...)
		}
	default: // leaf
		if conf == "[REQUIRED]" {
			path := strings.Join(fullPath, ".")
			errs = append(errs, errors.New(fmt.Sprintf("%s is reqired! Please set it by adding key -k %s:<value>", path, path)))
		}
	}
	return
}
