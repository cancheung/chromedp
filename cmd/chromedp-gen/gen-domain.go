// +build ignore

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"sort"
)

var (
	flagFile = flag.String("file", "protocol.json", "path to protocol.json")
	flagOut  = flag.String("out", "internal/domain.go", "out file")
)

func main() {
	var v struct {
		Domains []struct {
			Domain string `json:"domain"`
		} `json:"domains"`
	}

	buf, err := ioutil.ReadFile(*flagFile)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(buf, &v)
	if err != nil {
		log.Fatal(err)
	}

	var domains []string
	for _, d := range v.Domains {
		domains = append(domains, d.Domain)
	}
	sort.Strings(domains)

	var a, b string
	for _, s := range domains {
		a += fmt.Sprintf("Domain%s DomainType = \"%s\"\n", s, s)
		b += fmt.Sprintf("case Domain%s:\n*dt = Domain%s\n", s, s)
	}

	buf = []byte(fmt.Sprintf(tpl, a, b))
	err = ioutil.WriteFile(*flagOut, buf, 0644)
	if err != nil {
		log.Fatal(err)
	}

	err = exec.Command("gofmt", "-w", "-s", *flagOut).Run()
	if err != nil {
		log.Fatal(err)
	}
}

const (
	tpl = `package internal

// Code generated by gen-domain.go. DO NOT EDIT.

import (
	"fmt"
	"strconv"
)

// DomainType is the Chrome domain type.
type DomainType string

// DomainType values.
const (
%s)

// String satisfies Stringer.
func (dt DomainType) String() string {
	return string(dt)
}

// MarshalJSON satisfies json.Marshaler.
func (dt DomainType) MarshalJSON() ([]byte, error) {
	return []byte("\"" + dt + "\""), nil
}

// UnmarshalJSON satisfies json.Unmarshaler.
func (dt *DomainType) UnmarshalJSON(buf []byte) error {
	s, err := strconv.Unquote(string(buf))
	if err != nil {
		return err
	}

	switch DomainType(s) {
%s
	default:
		return fmt.Errorf("unknown domain type %%s", string(buf))
	}

	return nil
}
`
)
