package main

import (
    "log"
    "flag"
    "strings"
    "fmt"
    "deployer/libdeploy"
)

type replacers []string

func (i *replacers) String() string {
    return fmt.Sprint(*i)
}

func (i *replacers) Set(value string) error {
    for _, val := range strings.Split(value,",") {
        *i = append(*i, val)
    }
    return nil
}

func main() {
    log.Println("Hello!")
    var k replacers
    flag.Var(&k, "k", "Fields to replace")
    flag.Parse()
    var fn = flag.Args()

    config := libdeploy.ReadConfig(fn[0])

    for _, rep := range k {
        libdeploy.PutVariable(rep, config)
    }

   libdeploy.PrintConfig(config) 
}
