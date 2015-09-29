package zhash

import (
	"fmt"
	"strings"
)

// Retrieves []interface{} from hash. Will fail if target slice have different
// type ([]int for example).
func (h Hash) GetSlice(path ...string) ([]interface{}, error) {
	m := h.Get(path...)
	if m == nil {
		return []interface{}{}, notFoundError{path}
	}
	switch val := m.(type) {
	case []interface{}:
		return val, nil
	default:
		return []interface{}{}, fmt.Errorf(
			"cannot convert %s to slice", strings.Join(path, "."),
		)
	}
}

// Returns []int64 if any of []int, []int64 or []interface{} is found under the
// path. If target is []interface{} it will fails to convert if type of any
// element is not int or int64.
func (h Hash) GetIntSlice(path ...string) ([]int64, error) {
	m := h.Get(path...)
	if m == nil {
		return []int64{}, notFoundError{path}
	}
	switch val := m.(type) {
	case []int:
		sl := []int64{}
		for _, v := range val {
			sl = append(sl, int64(v))
		}
		return sl, nil
	case []int64:
		return val, nil
	case []interface{}:
		sl := []int64{}
		for _, v := range val {
			switch i := v.(type) {
			case int:
				sl = append(sl, int64(i))
			case int64:
				sl = append(sl, i)
			default:
				return []int64{}, fmt.Errorf(
					"cannot convert %s to []int64, "+
						"slice have not int elements",
					strings.Join(path, "."),
				)
			}
		}
		return sl, nil
	default:
		return []int64{}, fmt.Errorf(
			"cannot convert %s to []int64", strings.Join(path, "."),
		)
	}
}

func (h Hash) GetFloatSlice(path ...string) ([]float64, error) {
	m := h.Get(path...)
	if m == nil {
		return []float64{}, notFoundError{path}
	}
	switch val := m.(type) {
	case []float64:
		return val, nil
	case []interface{}:
		sl := []float64{}
		for _, v := range val {
			switch f := v.(type) {
			case float64:
				sl = append(sl, f)
			default:
				return []float64{}, fmt.Errorf(
					"cannot convert %s to []float64, "+
						"slice have not float elements",
					strings.Join(path, "."),
				)
			}
		}
		return sl, nil
	default:
		return []float64{}, fmt.Errorf(
			"cannot convert %s []float64", strings.Join(path, "."),
		)
	}
}

func (h Hash) GetStringSlice(path ...string) ([]string, error) {
	m := h.Get(path...)
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
				return []string{}, fmt.Errorf(
					"cannot convert %s to []string, "+
						"slice have not string elements",
					strings.Join(path, "."),
				)
			}
		}
		return sl, nil
	default:
		return []string{}, fmt.Errorf(
			"cannot convert %s []string", strings.Join(path, "."),
		)
	}
}

func (hash Hash) GetMapSlice(path ...string) ([]map[string]interface{}, error) {
	node := hash.Get(path...)
	if node == nil {
		return []map[string]interface{}{}, notFoundError{path}
	}
	result := []map[string]interface{}{}
	for _, elem := range node.([]interface{}) {
		switch typedElem := elem.(type) {
		case map[string]interface{}:
			result = append(result, typedElem)
		case map[interface{}]interface{}:
			newMap := make(map[string]interface{})
			for key, value := range typedElem {
				if newKey, ok := key.(string); ok {
					newMap[newKey] = value
				}
			}
			result = append(result, newMap)
		default:
			// do nothing
		}
	}
	return result, nil
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

func (h Hash) AppendIntSlice(val int64, path ...string) error {
	slice, err := h.GetIntSlice(path...)
	if err != nil {
		if !IsNotFound(err) {
			return err
		}
	}

	slice = append(slice, val)
	h.Set(slice, path...)
	return nil
}

func (h Hash) AppendFloatSlice(val float64, path ...string) error {
	slice, err := h.GetFloatSlice(path...)
	if err != nil {
		if !IsNotFound(err) {
			return err
		}
	}

	slice = append(slice, val)

	h.Set(slice, path...)
	return nil
}

func (h Hash) AppendStringSlice(val string, path ...string) error {
	slice, err := h.GetStringSlice(path...)
	if err != nil {
		if !IsNotFound(err) {
			return err
		}
	}

	slice = append(slice, val)

	h.Set(slice, path...)
	return nil
}

func (hash Hash) AppendMapSlice(val map[string]interface{}, path ...string) error {
	slice, err := hash.GetMapSlice(path...)
	if err != nil && !IsNotFound(err) {
		return err
	}

	slice = append(slice, val)

	hash.Set(slice, path...)
	return nil
}
