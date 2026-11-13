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

	var (
		report acctest.CompressReport
		err    error
	)

	report, err = acctest.CompressCassetteV4(path, true)
	if err != nil {
		log.Fatalf("%s", err)
	}

	report.Print()
}
