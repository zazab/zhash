package zhash_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

	"github.com/zazab/zhash"
)

// Example HashFromMap shows how to initialize your hash from map
func Example_hashFromMap() {
	m := map[string]interface{}{
		"plainValue": 10.1,
		"subMap": map[string]interface{}{
			"elem1": 10,
			"elem2": true,
		},
	}

	h := zhash.HashFromMap(m)

	f, err := h.GetFloat("plainValue")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(f)
	// Output:
	// 10.1
}

// Example ReadHash shows how to initialize your hash using Unmarshal function
// You can use any function that satisfies Unmarshaller type for ReadHash. For
// example, see TomlExample for example of using BurntSushi/toml for
// unmarshalling
func Example_readHash() {
	h := zhash.NewHash()
	h.SetMarshallerFunc(json.Marshal)
	h.SetUnmarshallerFunc(json.Unmarshal)

	b := bytes.NewBuffer([]byte("{\"this\": \"is\", \"some_json_map\": 14.1}"))

	err := h.ReadHash(b)
	if err != nil {
		log.Fatal(err)
	}

	s, err := h.GetString("this")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("This %s working", s)
	// Output:
	// This is working
}
