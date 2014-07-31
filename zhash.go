package zhash

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/BurntSushi/toml"
)

const (
	REQUIRED    = "[REQUIRED]"
	TIME_FORMAT = "2006-01-02T15:04:05Z"
)

type Hash struct {
	data map[string]interface{}
}

func NewHash() Hash {
	return Hash{map[string]interface{}{}}
}

type NotFoundError struct {
	Path []string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("Value for %s not found", strings.Join(e.Path, "."))
}

func NewNotFoundError(path []string) error {
	return NotFoundError{path}
}

type RequiredError struct {
	Path string
}

func (e RequiredError) Error() string {
	return fmt.Sprintf("%s is required, please specify it by adding "+
		"key -k %s:<value>", e.Path, e.Path)
}

func (c *Hash) ReadHash(r io.Reader) error {
	_, err := toml.DecodeReader(r, &c.data)
	return err
}

func (c Hash) WriteHash(w io.Writer) error {
	return toml.NewEncoder(w).Encode(c.data)
}

func (c Hash) Reader() io.Reader {
	var buff bytes.Buffer
	c.WriteHash(&buff)
	return &buff
}

func (c Hash) SetPath(value interface{}, path string) {
	c.Set(value, strings.Split(path, ".")...)
}

func (c Hash) Set(value interface{}, path ...string) {
	key := ""
	ptr := map[string]interface{}(c.data)
	for i, p := range path {
		if i < len(path)-1 { // middle element
			switch node := ptr[p].(type) {
			case map[string]interface{}:
				ptr = node
			default:
				ptr[p] = map[string]interface{}{}
				ptr = ptr[p].(map[string]interface{})
			}
		}
		key = p
	}

	ptr[key] = value
}

func (c Hash) GetPath(path ...string) interface{} {
	ptr := c.data
	for i, p := range path {
		if i == len(path)-1 {
			return ptr[p]
		}

		switch node := ptr[p].(type) {
		case map[string]interface{}:
			ptr = node
		default:
			return nil
		}
	}

	return nil
}

func (c Hash) GetMap(path ...string) (map[string]interface{}, error) {
	m := c.GetPath(path...)
	if m == nil {
		return map[string]interface{}{}, NewNotFoundError(path)
	}
	switch val := m.(type) {
	case map[string]interface{}:
		return val, nil
	default:
		return map[string]interface{}{},
			errors.New(fmt.Sprintf("Error converting %s to map",
				strings.Join(path, ".")))
	}
}

func (c Hash) GetString(path ...string) (string, error) {
	m := c.GetPath(path...)
	if m == nil {
		return "", NewNotFoundError(path)
	}
	switch val := m.(type) {
	case string:
		return val, nil
	default:
		return "", errors.New(fmt.Sprintf("Error converting %s to string",
			strings.Join(path, ".")))
	}
}

func (c Hash) GetSlice(path ...string) ([]interface{}, error) {
	m := c.GetPath(path...)
	if m == nil {
		return []interface{}{}, NewNotFoundError(path)
	}
	switch val := m.(type) {
	case []interface{}:
		return val, nil
	default:
		return []interface{}{},
			errors.New(fmt.Sprintf("Error converting %s to slice",
				strings.Join(path, ".")))
	}
}

func (c Hash) GetStringSlice(path ...string) ([]string, error) {
	m := c.GetPath(path...)
	if m == nil {
		return []string{}, NewNotFoundError(path)
	}
	switch val := m.(type) {
	case []interface{}:
		sl := []string{}
		for _, v := range val {
			switch s := v.(type) {
			case string:
				sl = append(sl, s)
			default:
				return []string{}, errors.New(
					fmt.Sprintf("Error converting %s to string slice",
						strings.Join(path, ".")))
			}
		}
		return sl, nil
	default:
		return []string{},
			errors.New(fmt.Sprintf("Error converting %s to slice",
				strings.Join(path, ".")))
	}
}

func (c Hash) GetBool(path ...string) (bool, error) {
	m := c.GetPath(path...)
	if m == nil {
		return false, NewNotFoundError(path)
	}
	switch val := m.(type) {
	case bool:
		return val, nil
	default:
		return false, errors.New(fmt.Sprintf("Error converting %s to bool",
			strings.Join(path, ".")))
	}
}

func (c Hash) GetInt(path ...string) (int64, error) {
	m := c.GetPath(path...)
	if m == nil {
		return 0, NewNotFoundError(path)
	}
	switch val := m.(type) {
	case int:
		return int64(val), nil
	case int64:
		return val, nil
	default:
		return 0, errors.New(fmt.Sprintf("Error converting %s to int",
			strings.Join(path, ".")))
	}
}

func (c Hash) GetFloat(path ...string) (float64, error) {
	m := c.GetPath(path...)
	if m == nil {
		return 0, NewNotFoundError(path)
	}
	switch val := m.(type) {
	case float64:
		return val, nil
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	default:
		return 0, errors.New(fmt.Sprintf("Error converting %s to float",
			strings.Join(path, ".")))
	}
}

func (c Hash) Validate() (errs []error) {
	nodes := []interface{}{}
	paths := []string{}

	for p, v := range c.data {
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
			if inner == REQUIRED {
				errs = append(errs, RequiredError{path})
			}
		}
	}

	return
}

func (c Hash) String() string {
	buf, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return "Error converting config to json"
	}

	return string(buf)
}
