package main

import (
	"log"
	"os"

	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <cassette_file_name_without_yaml>\n", os.Args[0])
	}

	path := os.Args[1]

	report, err := acctest.CompressCassette(path)
	if err != nil {
		panic(err)
	}

	report.Print()
}
