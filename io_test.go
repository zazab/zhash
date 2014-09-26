package zhash

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"testing"
)

func TestReadHashSuccess(t *testing.T) {
	fd, err := os.Open("test.json")
	if err != nil {
		t.Error("Cannot read test.json!")
	}
	defer fd.Close()

	h := NewHash()
	h.SetUnmarshallerFunc(json.Unmarshal)
	err = h.ReadHash(fd)
	if err != nil {
		t.Error("Error reading hash!")
	}
}

func TestReadHashFailParse(t *testing.T) {
	fd, err := os.Open("test_corrupted.json")
	if err != nil {
		t.Error("Cannot read test_corrupted.json!")
	}
	defer fd.Close()

	h := NewHash()
	h.SetUnmarshallerFunc(json.Unmarshal)
	err = h.ReadHash(fd)
	if err == nil {
		t.Error("Hash readed from corrupted source!")
	}
}

func TestReadHashFailNoUnmarshaller(t *testing.T) {
	fd, err := os.Open("test_corrupted.json")
	if err != nil {
		t.Error("Cannot read test_corrupted.json!")
	}
	defer fd.Close()

	h := NewHash()
	err = h.ReadHash(fd)
	if err == nil {
		t.Error("Hash readed from corrupted source!")
	}
}

type corruptedReader string

func (c corruptedReader) Read(b []byte) (int, error) {
	return 0, errors.New("I'm corrupted!")
}

func TestReadHashFailReaderErr(t *testing.T) {
	r := corruptedReader("aaA")

	h := NewHash()
	h.SetUnmarshallerFunc(json.Unmarshal)
	err := h.ReadHash(r)
	if err == nil {
		t.Error("Hash readed from corrupted source!")
	}
}

func TestHashReader(t *testing.T) {
	hash := HashFromMap(testMap)
	hash.SetMarshallerFunc(json.Marshal)
	r, err := hash.Reader()
	if err != nil {
		t.Error("Error marshalling Hash")
	}
	f, err := os.OpenFile(os.DevNull, os.O_RDWR, 0666)
	if err != nil {
		t.Error("Error opening DevNull:", err)
	}
	defer f.Close()
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		t.Error("Error reading from hash.Reader():", err)
	}
	f.Write(buf)
}

func TestToStringSuccess(t *testing.T) {
	hash := HashFromMap(testMap)

	t.Logf("Hash: %s", hash)
}

type buggyStruct struct {
	Id int `json:"id"`
}

func (b buggyStruct) MarshalJSON() ([]byte, error) {
	return nil, errors.New("Baka!")
}

func TestToStringFail(t *testing.T) {
	hash := NewHash()

	value := buggyStruct{Id: 10}
	hash.Set(value, "meta", "bug")

	t.Logf("Hash: %s", hash)
}

func TestToJson(t *testing.T) {
	hash := HashFromMap(map[string]interface{}{
		"rec1": "val one",
		"rec2": map[string]interface{}{
			"sub_rec1": 2,
			"sub_rec2": "string",
		},
	})

	jsonText := "{\"rec1\":\"val one\",\"rec2\":{\"sub_rec1\":2,\"sub_rec2\":\"string\"}}"

	convert, err := json.Marshal(hash)
	if err != nil {
		t.Error("Error marshalling hash to json:", err)
	}

	if string(convert) != jsonText {
		t.Errorf("Marshalled json differs from wanted:\nWant: %s\nGot: %s",
			jsonText, string(convert))
	}
}

func TestWriteHash(t *testing.T) {
	hash := HashFromMap(testMap)
	hash.SetMarshallerFunc(json.Marshal)
	f, _ := os.OpenFile(os.DevNull, os.O_RDWR, 0666)
	defer f.Close()
	if err := hash.WriteHash(f); err != nil {
		t.Errorf("Errors while writing hashig: %s", err.Error())
	}
}

func TestWriteHashError(t *testing.T) {
	hash := NewHash()
	hash.SetMarshallerFunc(json.Marshal)
	hash.Set(buggyStruct{10}, "bug")
	f, _ := os.OpenFile(os.DevNull, os.O_RDWR, 0666)
	defer f.Close()
	if err := hash.WriteHash(f); err == nil {
		t.Errorf("No error while marshalling buggyStruct!")
	}
}

func TestWriteHashNoMarshaller(t *testing.T) {
	hash := HashFromMap(testMap)
	f, _ := os.OpenFile(os.DevNull, os.O_RDWR, 0666)
	defer f.Close()
	if err := hash.WriteHash(f); err == nil {
		t.Errorf("WriteHash doesn't return any error, but should")
	}
}
