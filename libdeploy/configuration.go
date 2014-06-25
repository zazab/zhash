package configuration

import (
	"bytes"
	"errors"
	"github.com/BurntSushi/toml"
	"io"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"
)

type Config map[string]interface{}

var timeFormat = "2006-01-02T15:04:05Z"

func (c *Config) ReadConfig(r io.Reader) (err error) {
	var buffer []byte
	buffer, err = ioutil.ReadAll(r)
	if err != nil {
		return
	}

	_, err = toml.Decode(string(buffer), &c)

	return
}

func (c Config) WriteConfig(w io.Writer) (err error) {
	buf := new(bytes.Buffer)
	if err = toml.NewEncoder(buf).Encode(c); err != nil {
		return
	}

	w.Write(buf.Bytes())
	return
}

func (c Config) SetVariable(path string, value interface{}) {
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

func (c Config) ReplaceConfigParameter(path string) {
	buf := strings.SplitN(path, ":", 2)
	path = buf[0]
	val := buf[1]

	if t, err := time.Parse(timeFormat, val); err != nil {
		if i, err := strconv.Atoi(val); err != nil {
			if r, err := strconv.ParseFloat(val, 64); err != nil {
				if b, err := strconv.ParseBool(val); err != nil {
					c.SetVariable(path, val) // Cannot conver to any type, sujesting string
				} else { // Converted to bool
					c.SetVariable(path, b)
				}
			} else { // Converted to float
				c.SetVariable(path, r)
			}
		} else { // Converted to int
			c.SetVariable(path, i)
		}
	} else { // Converted to time
		c.SetVariable(path, t)
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
