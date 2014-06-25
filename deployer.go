package main

import (
	"deployer/libdeploy"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

type replacers []string

func (i *replacers) String() string {
	return fmt.Sprint(*i)
}

func (i *replacers) Set(value string) error {
	for _, val := range strings.Split(value, ",") {
		*i = append(*i, val)
	}
	return nil
}

func main() {
	var k replacers
	flag.Var(&k, "k", "Fields to replace")
	flag.Parse()
	var fn = flag.Args()
	var config configuration.Config

	fd, err := os.Open(fn[0])
	if err != nil {
		log.Fatal("Profile not found")
	}
	defer fd.Close()
	err = config.ReadConfig(fd)
	if err != nil {
		log.Fatal("Cannot read config: ", err)
		return
	}

	for _, rep := range k {
		config.ReplaceConfigParameter(rep)
	}

	errs := config.Validate()
	if errs != nil {
		for _, err := range errs {
			log.Println(err)
		}
		log.Fatal("Missing required arguments")
	}

	config.WriteConfig(os.Stdout)
}
