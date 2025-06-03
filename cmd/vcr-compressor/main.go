package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/scaleway/scaleway-sdk-go/api/k8s/v1"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
)

var transientStates = map[string]bool{
	k8s.ClusterStatusCreating.String(): true,
	k8s.ClusterStatusDeleting.String(): true,
	k8s.ClusterStatusUpdating.String(): true,
	k8s.PoolStatusDeleting.String():    true,
	k8s.PoolStatusScaling.String():     true,
	k8s.PoolStatusUpgrading.String():   true,
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <cassette_file_name_without_yaml>\n", os.Args[0])
	}

	path := os.Args[1]

	inputCassette, err := cassette.Load(path)
	if err != nil {
		log.Fatalf("Error while reading file : %v\n", err)
	}

	outputCassette := cassette.New(path)
	transitioning := false

	for i := range len(inputCassette.Interactions) {
		interaction := inputCassette.Interactions[i]
		responseBody := interaction.Response.Body
		requestMethod := interaction.Request.Method

		if requestMethod != "GET" {
			transitioning = false

			log.Printf("Interaction %d is not a GET request. Recording it\n", i)
			outputCassette.AddInteraction(interaction)

			continue
		}

		if responseBody == "" {
			log.Printf("Interaction %d got an empty response body. Recording it\n", i)
			outputCassette.AddInteraction(interaction)

			continue
		}

		var m map[string]interface{}

		err := json.Unmarshal([]byte(responseBody), &m)
		if err != nil {
			log.Printf("Interaction %d have an error with unmarshalling response body: %v. Recording it\n", i, err)
			outputCassette.AddInteraction(interaction)

			continue
		}

		if m["status"] == nil {
			log.Printf("Interaction %d does not contain a status field. Recording it\n", i)
			outputCassette.AddInteraction(interaction)

			continue
		}

		status := m["status"].(string)
		// We test if the state is transient
		if _, ok := transientStates[status]; ok {
			if transitioning {
				log.Printf("Interaction %d is in a transient state while we are already in transitient state. No need to record it: %s\n", i, status)

				continue
			} else {
				log.Printf("Interaction %d is in a transient state: %s, Recording it\n", i, status)

				transitioning = true

				outputCassette.AddInteraction(interaction)
			}
		} else {
			if transitioning {
				log.Printf("Interaction %d is not in a transient state anymore: %s, Recording it\n", i, status)

				outputCassette.AddInteraction(interaction)

				transitioning = false
			} else {
				log.Printf("Interaction %d is not in a transient state: %s, Recording it\n", i, status)
				outputCassette.AddInteraction(interaction)
			}
		}
	}

	err = outputCassette.Save()
	if err != nil {
		log.Fatalf("error while saving file: %v", err)
	}
}
