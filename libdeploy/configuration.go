package libdeploy

import (
	"bytes"
	"fmt"
	"github.com/BurntSushi/toml"
	"io"
	"strings"
)

type Config map[string]interface{}

type ErrorRequired struct {
	Path string
}

func (e ErrorRequired) Error() string {
	return fmt.Sprintf("%s is required, please specify it by adding key -k %s:<value>", e.Path, e.Path)
}

func (c *Config) ReadConfig(r io.Reader) (err error) {
	_, err = toml.DecodeReader(r, &c)
	return
}

func (c Config) WriteConfig(w io.Writer) (err error) {
	err = toml.NewEncoder(w).Encode(c)
	return
}

func (c Config) Reader() io.Reader {
	var buff bytes.Buffer
	c.WriteConfig(&buff)
	return &buff
}

func (c Config) SetVariable(path string, value interface{}) {
	key := ""
	ptr := map[string]interface{}(c)
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

func (c Config) GetMap(path ...string) map[string]interface{} {
	m := c.GetPath(path...)
	if m == nil {
		return nil
	}
	return m.(map[string]interface{})
}

func (c Config) GetString(path ...string) string {
	m := c.GetPath(path...)
	if m == nil {
		return ""
	}
	return m.(string)
}

func (c Config) GetSlice(path ...string) []interface{} {
	m := c.GetPath(path...)
	if m == nil {
		return []interface{}{}
	}
	return m.([]interface{})
}

func (c Config) GetInt(path ...string) int {
	m := c.GetPath(path...)
	if m == nil {
		return 0
	}
	return m.(int)
}

func (c Config) GetFloat(path ...string) float64 {
	m := c.GetPath(path...)
	if m == nil {
		return 0
	}
	return m.(float64)
}

func (c Config) Validate() (errs []error) {
	nodes := []interface{}{}
	paths := []string{}

	for p, v := range c {
		nodes = append(nodes, v)
		paths = append(paths, p)
	}

	for len(nodes) > 0 {
		node := nodes[len(nodes)-1]
		path := paths[len(paths)-1]
		nodes = nodes[:len(nodes)-1]
		paths = paths[:len(paths)-1]

		switch inner := node.(type) {
		case map[string]interface{}:
			for k, n := range inner {
				nodes = append(nodes, n)
				paths = append(paths, path+"."+k)
			}
		case string:
			if inner == "[REQUIRED]" {
				errs = append(errs, ErrorRequired{path})
			}
		}
	}

	return
}
