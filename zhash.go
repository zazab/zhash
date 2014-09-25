package zhash

import (
	"errors"
	"fmt"
	"log"
	"strings"
)

type Hash struct {
	data      map[string]interface{}
	marshal   Marshaller
	unmarshal Unmarshaller
}

func NewHash(m Marshaller, u Unmarshaller) Hash {
	return Hash{map[string]interface{}{}, m, u}
}

func HashFromMap(ma map[string]interface{}, m Marshaller, u Unmarshaller) Hash {
	return Hash{ma, m, u}
}

type notFoundError struct {
	path []string
}

func (e notFoundError) Error() string {
	return fmt.Sprintf("Value for %s not found", strings.Join(e.path, "."))
}

func IsNotFound(err error) bool {
	_, ok := err.(notFoundError)
	return ok
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

func (h Hash) AppendSlice(val interface{}, path ...string) error {
	slice, err := h.GetSlice(path...)
	if err != nil {
		if !IsNotFound(err) {
			return err
		}
	}

	slice = append(slice, val)

	h.Set(slice, path...)
	return nil
}

func (h Hash) AppendStringSlice(val string, path ...string) {
	slice, err := h.GetStringSlice(path...)
	if err != nil {
		log.Println("Error getting slice:", err)
	}

	slice = append(slice, val)

	h.Set(slice, path...)
}

func (h Hash) Delete(path ...string) error {
	l := len(path)
	if l == 1 {
		delete(h.data, path[0])
		return nil
	}

	elemPath := path[l-1]
	parentPath := path[:l-1]
	parent := h.GetPath(parentPath...)

	if parent == nil {
		return notFoundError{path}
	}

	switch val := parent.(type) {
	case map[string]interface{}:
		delete(val, elemPath)
		return nil
	default:
		errmsg := fmt.Sprintf("Cannot delete key %s from %T, "+
			"expected map[string]interface{}", parent)
		return errors.New(errmsg)
	}
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
		return map[string]interface{}{}, notFoundError{path}
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
		return "", notFoundError{path}
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
		return []interface{}{}, notFoundError{path}
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
		return []string{}, notFoundError{path}
	}
	switch val := m.(type) {
	case []string:
		return val, nil
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
		return false, notFoundError{path}
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
		return 0, notFoundError{path}
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
		return 0, notFoundError{path}
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
