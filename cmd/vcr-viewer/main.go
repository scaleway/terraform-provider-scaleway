package main

import (
	"encoding/json"
	"log"
	"os"

	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <cassette_file_name_without_yaml>\n", os.Args[0])
	}

	path := os.Args[1]

	data, err := cassette.Load(path)
	if err != nil {
		log.Fatalf("Error while reading file: %v\n", err)
	}

	for i := range len(data.Interactions) {
		interaction := data.Interactions[i]

		log.Println("--------------")
		log.Printf("Interaction %d:\n", i+1)
		log.Printf("  Request:\n")
		log.Printf("    Method: %s\n", interaction.Request.Method)
		log.Printf("    URL: %s\n", interaction.Request.URL)

		if interaction.Request.Body != "" {
			log.Printf("    Body: %s\n", interaction.Request.Body)
		}

		log.Printf("  Response:\n")
		log.Printf("    Status: %s\n", interaction.Response.Status)
		log.Printf("    Body: %s\n", interaction.Response.Body)

		var m map[string]any

		err := json.Unmarshal([]byte(interaction.Response.Body), &m)
		if err != nil {
			continue
		}

		if m["status"] != nil {
			log.Println("++++++++++++++++")
			log.Printf("status: %s\n", m["status"])
			log.Println("++++++++++++++++")
		}

		log.Println("--------------")
	}
}
