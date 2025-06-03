package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"log"
	"os"
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

	for i := 0; i < len(data.Interactions); i++ {
		interaction := data.Interactions[i]
		fmt.Println("--------------")
		fmt.Printf("Interaction %d:\n", i+1)
		fmt.Printf("  Request:\n")
		fmt.Printf("    Method: %s\n", interaction.Request.Method)
		fmt.Printf("    URL: %s\n", interaction.Request.URL)
		if interaction.Request.Body != "" {
			fmt.Printf("    Body: %s\n", interaction.Request.Body)
		}
		fmt.Printf("  Response:\n")
		fmt.Printf("    Status: %s\n", interaction.Response.Status)
		fmt.Printf("    Body: %s\n", interaction.Response.Body)

		var m map[string]interface{}
		err := json.Unmarshal([]byte(interaction.Response.Body), &m)
		if err != nil {
			continue
		}
		if m["status"] != nil {
			fmt.Println("++++++++++++++++")
			fmt.Printf("status: %s\n", m["status"]) // Modifie le champ "status" pour qu'il soit "ok"
			fmt.Println("++++++++++++++++")
		}
		fmt.Println("--------------")
	}
}
