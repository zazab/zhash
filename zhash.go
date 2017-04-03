package zhash // import "github.com/zazab/zhash"

import (
	"fmt"
	"strings"
)

type Hash struct {
	data      map[string]interface{}
	marshal   Marshaller
	unmarshal Unmarshaller
}

func NewHash() Hash {
	return Hash{map[string]interface{}{}, nil, nil}
}

func NewHashPtr() *Hash {
	return &Hash{map[string]interface{}{}, nil, nil}
}

// Loads existing map[string]interface{} to Hash. Marshaller and Unmarshallers
// are optional, if you don't need it pass nil to them. You can set (or change)
// them later using Hash.SetMarshaller and Hash.SetUnmarshaller.
func HashFromMap(ma map[string]interface{}) Hash {
	return Hash{ma, nil, nil}
}

type notFoundError struct {
	path []string
}

func (e notFoundError) Error() string {
	return fmt.Sprintf("value for %s not found", strings.Join(e.path, "."))
}

// Check if given err represents zhash "Not Found" error. Great for checking if
// asked value is zero or just not set.
func IsNotFound(err error) bool {
	_, ok := err.(notFoundError)
	return ok
}

func (h Hash) Set(value interface{}, path ...string) {
	key := ""
	ptr := h.data
	for i, p := range path {
		if i < len(path)-1 { // middle element
			switch node := ptr[p].(type) {
			case map[string]interface{}:
				ptr = node
			case map[interface{}]interface{}:
				// golang yaml implementations parses data into
				// map[interface{}]interface{}. zhash works with
				// map[string]interface{}. So we need to  convert
				// map[interface{}]interface{} to map[string]interface{}
				// or it would be overwritten by empty map[string]interface{}
				// on set attempt
				ptr[p] = convertToMapString(node)
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

func (h *Hash) SetRoot(value map[string]interface{}) {
	h.data = value
}

func (h Hash) Delete(path ...string) error {
	l := len(path)
	if l == 1 {
		delete(h.data, path[0])
		return nil
	}

	elemPath := path[l-1]
	parentPath := path[:l-1]
	parent := h.Get(parentPath...)

	if parent == nil {
		return notFoundError{path}
	}

	switch val := parent.(type) {
	case map[string]interface{}:
		delete(val, elemPath)
		return nil
	default:
		return fmt.Errorf(
			"cannot delete key %s from %T, "+
				"expected map[string]interface{}",
			strings.Join(path, "."), parent,
		)
	}
}

// Retrieves value from hash returns nil if nothing found
func (h Hash) Get(path ...string) interface{} {
	ptr := h.data
	for i, p := range path {
		if i == len(path)-1 {
			if node, ok := ptr[p].(map[interface{}]interface{}); ok {
				return convertToMapString(node)
			}
			return ptr[p]
		}

		switch node := ptr[p].(type) {
		case map[string]interface{}:
			ptr = node
		case map[interface{}]interface{}:
			ptr = convertToMapString(node)
		default:
			return nil
		}
	}

	return nil
}

func convertToMapString(node map[interface{}]interface{}) map[string]interface{} {
	convertedNode := make(map[string]interface{})
	for key, val := range node {
		if keystr, ok := key.(string); ok {
			convertedNode[keystr] = val
		}
	}
	return convertedNode
}

// Returns root map[string]interface{}
func (h Hash) GetRoot() map[string]interface{} {
	return h.data
}

// Retrieves map[string]interface{} returns error if any can not convert
// target value, or value doesn't found. If not found, returns empty
// Hash, not nil
func (h Hash) GetMap(path ...string) (map[string]interface{}, error) {
	m := h.Get(path...)
	if m == nil {
		return map[string]interface{}{}, notFoundError{path}
	}
	switch val := m.(type) {
	case map[string]interface{}:
		return val, nil
	default:
		return map[string]interface{}{}, fmt.Errorf(
			"cannot convert %s to map", strings.Join(path, "."),
		)
	}
}

// Retrieves map[string]interface{} and converts it to Hash. Returns error if
// can not convert target value, or value doesn'n found. If not found returns
// emty map[string]interface{} not nil
func (h Hash) GetHash(path ...string) (Hash, error) {
	m := h.Get(path...)
	if m == nil {
		return NewHash(), notFoundError{path}
	}
	switch val := m.(type) {
	case map[string]interface{}:
		return HashFromMap(val), nil
	default:
		return NewHash(), fmt.Errorf(
			"cannot convert %s to map", strings.Join(path, "."),
		)
	}
}

// Returns root keys of Hash
func (h Hash) Keys() []string {
	keys := make([]string, len(h.data))
	i := 0
	for k, _ := range h.data {
		keys[i] = k
		i++
	}

	return keys
}

// Returns len of root map
func (h Hash) Len() int {
	return len(h.data)
}

func (h Hash) GetString(path ...string) (string, error) {
	m := h.Get(path...)
	if m == nil {
		return "", notFoundError{path}
	}
	switch val := m.(type) {
	case string:
		return val, nil
	default:
		return "", fmt.Errorf(
			"cannot convert %s to string", strings.Join(path, "."),
		)
	}
}

func (h Hash) GetBool(path ...string) (bool, error) {
	m := h.Get(path...)
	if m == nil {
		return false, notFoundError{path}
	}
	switch val := m.(type) {
	case bool:
		return val, nil
	default:
		return false, fmt.Errorf(
			"cannot convert %s to bool", strings.Join(path, "."),
		)
	}
}

func (h Hash) GetInt(path ...string) (int64, error) {
	m := h.Get(path...)
	if m == nil {
		return 0, notFoundError{path}
	}
	switch val := m.(type) {
	case int:
		return int64(val), nil
	case int64:
		return val, nil
	default:
		return 0, fmt.Errorf(
			"cannot convert %s to int", strings.Join(path, "."),
		)
	}
}

func (h Hash) GetFloat(path ...string) (float64, error) {
	m := h.Get(path...)
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
		return 0, fmt.Errorf(
			"cannot convert %s to float", strings.Join(path, "."),
		)
	}
}
