package main

import (
	"fmt"
	"log"

	"github.com/ecc1/cc111x"
)

func main() {
	r := cc111x.Open()
	if r.Error() != nil {
		log.Fatal(r.Error())
	}
	fmt.Printf("version: %s\n", r.Version())
	fmt.Printf("state: %s\n", r.State())
	fmt.Printf("old frequency: %d\n", r.Frequency())
	r.SetFrequency(900000000)
	fmt.Printf("new frequency: %d\n", r.Frequency())
}
