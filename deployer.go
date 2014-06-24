package main

import (
	"deployer/libdeploy"
	"flag"
	"fmt"
	"log"
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

	config, err := configuration.ReadConfig(fn[0])
	if err != nil {
		log.Fatal("Cannot read config: ", err)
		return
	}

	for _, rep := range k {
		if err := configuration.PutVariable(rep, config); err != nil {
			log.Fatal(fmt.Sprintf("Error during setting \"%s\":\t%+v", rep, err))
		}
	}

	errs := configuration.CheckRequired(config, nil)
	if errs != nil {
		for _, err := range errs {
			log.Println(err)
		}
		log.Fatal("Missing required arguments")
	}

	configuration.PrintConfig(config)
}
