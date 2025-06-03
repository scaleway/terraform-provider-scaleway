package main

import (
	"encoding/json"
	"log"
	"os"

	"gopkg.in/dnaeon/go-vcr.v3/cassette"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <chemin_du_fichier_yaml>\n", os.Args[0])
	}

	chemin := os.Args[1]

	data, err := cassette.Load(chemin)
	if err != nil {
		log.Fatalf("Erreur de lecture du fichier : %v\n", err)
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

		var m map[string]interface{}

		err := json.Unmarshal([]byte(interaction.Response.Body), &m)
		if err != nil {
			continue
		}

		if m["status"] != nil {
			log.Println("++++++++++++++++")
			log.Printf("status: %s\n", m["status"]) // Modifie le champ "status" pour qu'il soit "ok"
			log.Println("++++++++++++++++")
		}

		log.Println("--------------")
	}
}
