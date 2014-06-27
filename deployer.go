package main

import (
	"deployer/libdeploy"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

type replacments []string

func (i *replacments) String() string {
	return fmt.Sprint(*i)
}

func (i *replacments) Set(value string) error {
	for _, val := range strings.Split(value, ",") {
		*i = append(*i, val)
	}
	return nil
}

func main() {
	var k replacments
	flag.Var(&k, "k", "fields to replace")
	flag.Parse()
	var config libdeploy.Config

	fd, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal("Profile not found")
	}
	defer fd.Close()
	err = config.ReadConfig(fd)
	if err != nil {
		log.Fatal("Cannot read config:", err)
	}

	for _, rep := range k {
		config.SetVariable(libdeploy.ParseSetArgument(rep))
	}

	errs := config.Validate()
	if errs != nil {
		for _, err := range errs {
			log.Println(err.Error())
		}
		log.Fatal("Missing required arguments")
	}

	config.WriteConfig(os.Stdout)
}
