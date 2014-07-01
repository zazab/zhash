package main

import (
	"deployer/libdeploy"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

type replacements []string

func (i *replacements) String() string {
	return fmt.Sprint(*i)
}

func (i *replacements) Set(value string) error {
	*i = append(*i, strings.Split(value, ",")...)
	return nil
}

func main() {
	var k replacements
	flag.Var(&k, "k", "fields to replace")
	flag.Parse()
	config := libdeploy.NewConfig()

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
		config.SetPath(libdeploy.ParseSetArgument(rep))
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
