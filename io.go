package zhash

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
)

type Unmarshaller func([]byte, interface{}) error
type Marshaller func(interface{}) ([]byte, error)

// Sets function for marshalling via Hash.WriteHash(fd)
func (h *Hash) SetMarshallerFunc(fu Marshaller) {
	h.marshal = fu
}

// Set function for unmarshalling via Hash.ReadHash
func (h *Hash) SetUnmarshallerFunc(fu Unmarshaller) {
	h.unmarshal = fu
}

// Unmarshall hash from given io.Reader using function setted via zhash.Hash.SetUnmarshaller
func (h *Hash) ReadHash(r io.Reader) error {
	if h.unmarshal == nil {
		return errors.New("cannot unmarshal, no unmarshaller set")
	}

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	err = h.unmarshal(b, &h.data)
	return err
}

// Mashall hash using supplied Marshaller function and writes it to w
func (h Hash) WriteHash(w io.Writer) error {
	if h.marshal == nil {
		return errors.New("cannot marshal hash, no marshaller set")
	}

	b, err := h.marshal(h.data)
	if err != nil {
		return err
	}

	_, err = w.Write(b)
	return err
}

func (h Hash) Reader() (io.Reader, error) {
	var buff bytes.Buffer
	err := h.WriteHash(&buff)
	return &buff, err
}

// Returns indented json with your map
func (h Hash) String() string {
	buf, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return "error converting config to json"
	}

	return string(buf)
}

func (h Hash) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.data)
}
