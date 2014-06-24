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

type Config map[string]interface{}

func ReadConfig(fileName string) (config Config, err error) {
	blob, err := ioutil.ReadFile(fileName)
	if err != nil {
		return
	}

	_, err = toml.Decode(string(blob), &config)

	return
}

func SPrintConfig(config Config) (conf string, err error) {
	buf := new(bytes.Buffer)
	if err = toml.NewEncoder(buf).Encode(config); err != nil {
		return
	}

	conf = buf.String()
	return
}

func PrintConfig(config Config) (err error) {
	buf := new(bytes.Buffer)
	if err = toml.NewEncoder(buf).Encode(config); err != nil {
		return
	}

	fmt.Println(buf.String())
	return
}

func PutVariable(path string, config Config) (err error) {
	var full_path = "config"
	var buffer, changer map[string]interface{}
	var last_path string

	buf := strings.Split(path, ":")
	// FIXME add check that only one semicolon is used
	path = buf[0]
	val := strings.Join(buf[1:], ":")
	variable_path := strings.Split(path, ".")
	for num, p := range variable_path {
		full_path = fmt.Sprintf("%s[%s]", full_path, p)
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

	if t, err := time.Parse(timeFormat, val); err != nil {
		if i, err := strconv.Atoi(val); err != nil {
			if r, err := strconv.ParseFloat(val, 64); err != nil {
				if b, err := strconv.ParseBool(val); err != nil {
					changer[last_path] = val // Cannot conver to any type, sujesting string
				} else { // Converted to bool
					changer[last_path] = b
				}
			} else { // Converted to float
				changer[last_path] = r
			}
		} else { // Converted to int
			changer[last_path] = i
		}
	} else { // Converted to time
		changer[last_path] = t
	}

	return
}

func CheckRequired(conf interface{}, fullPath []string) (errs []error) {
	switch conf.(type) {
	case Config:
		for p, val := range conf.(Config) {
			errs = append(errs, CheckRequired(val, append(fullPath, p))...)
		}
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
