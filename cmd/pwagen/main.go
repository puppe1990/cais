package main

import (
	"log"
	"os"

	"github.com/matheuspuppe/cais/pkg/cais/pwa"
)

func main() {
	name := "Cais"
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	if err := pwa.InstallTo(".", name); err != nil {
		log.Fatal(err)
	}
}
