package libdeploy

import (
	"bytes"
	"github.com/BurntSushi/toml"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

type Config map[string]interface{}

type ErrorRequired struct {
	Path string
}

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
	var ptr map[string]interface{}
	var key string = ""

	ptr = c
	path_way := strings.Split(path, ".")
	for i, p := range path_way {
		if i < len(path_way)-1 { // middle element
			switch ptr[p].(type) {
			case map[string]interface{}:
				ptr = ptr[p].(map[string]interface{})
			default:
				ptr[p] = map[string]interface{}{}
				ptr = ptr[p].(map[string]interface{})
			}
		}
		key = p
	}

	ptr[key] = value
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

func (c Config) GetPath(path ...string) interface{} {
	ptr := c
	for i, p := range path {
		switch ptr[p].(type) {
		case map[string]interface{}:
			if i == len(path)-1 {
				return ptr[p]
			}
			ptr = ptr[p].(map[string]interface{})
		default:
			if i < len(path)-1 {
				return nil
			} else {
				return ptr[p]
			}
		}
	}

	return nil
}

type valNode struct {
	path  string
	value interface{}
}

func (c Config) Validate() (errs []ErrorRequired) {
	s := NewStack()

	for f, v := range c {
		s.Push(valNode{f, v})
	}

	for s.Len() > 0 {
		n := s.Pop().(valNode)
		switch n.value.(type) {
		case map[string]interface{}:
			for f, v := range n.value.(map[string]interface{}) {
				p := n.path + "." + f
				s.Push(valNode{p, v})
			}
		default:
			if n.value == "[REQUIRED]" {
				errs = append(errs, ErrorRequired{n.path})
			}
		}
	}

	return
}
