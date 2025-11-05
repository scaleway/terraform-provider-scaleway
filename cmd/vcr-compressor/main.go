package main

import (
	"log"
	"os"
	"strings"

	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <cassette_file_name_without_yaml>\n", os.Args[0])
	}

	path := os.Args[1]
	folder := strings.Split(path, "/")[2]

	var (
		report acctest.CompressReport
		err    error
	)

	if acctest.FolderUsesVCRv4(folder) {
		report, err = acctest.CompressCassetteV4(path)
	} else {
		report, err = acctest.CompressCassetteV3(path)
	}

	if err != nil {
		log.Fatalf("%s", err)
	}

	report.Print()
}
