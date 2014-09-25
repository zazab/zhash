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

func (h *Hash) SetMarshallerFunc(fu Marshaller) {
	h.marshal = fu
}

func (h *Hash) SetUnmarshallerFunc(fu Unmarshaller) {
	h.unmarshal = fu
}

func (c *Hash) ReadHash(r io.Reader) error {
	if c.unmarshal == nil {
		return errors.New("Cannot unmarshal. No unmarshaller set")
	}

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	err = c.unmarshal(b, &c.data)
	return err
}

func (c Hash) WriteHash(w io.Writer) error {
	if c.marshal == nil {
		return errors.New("Cannot marshal hash. No marshaller set")
	}

	b, err := c.marshal(c.data)
	if err != nil {
		return err
	}

	_, err = w.Write(b)
	return err
}

func (c Hash) Reader() io.Reader {
	var buff bytes.Buffer
	c.WriteHash(&buff)
	return &buff
}

func (c Hash) String() string {
	buf, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return "Error converting config to json"
	}

	return string(buf)
}

func (h Hash) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.data)
}
