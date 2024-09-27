package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate/entities/models"
)

var (
	WEAVIATE_URL  = os.Getenv("WEAVIATE_URL")  // "localhost:8080"
	VECTORIZER    = os.Getenv("VECTORIZER")    // "none", "text2vec-contextionary", "text2vec-openai"
	OPENAI_APIKEY = os.Getenv("OPENAI_APIKEY") // ""
)

func main() {
	validateEnvVars()

	client, err := createWeaviateClient()
	if err != nil {
		panic(err)
	}

	if err := instantiateSchema(client, VECTORIZER); err != nil {
		panic(err)
	}
}

func validateEnvVars() {
	if WEAVIATE_URL == "" {
		log.Fatal("WEAVIATE_URL environment variable is not set")
	}
	if VECTORIZER == "" {
		log.Fatal("VECTORIZER environment variable is not set")
	}
}

func createWeaviateClient() (*weaviate.Client, error) {
	cfg := weaviate.Config{
		Host:   WEAVIATE_URL,
		Scheme: "http",
		Headers: map[string]string{
			"X-OpenAI-Api-Key": OPENAI_APIKEY,
		},
	}

	client, err := weaviate.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	live, err := client.Misc().LiveChecker().Do(context.Background())
	fmt.Printf("Connected to Weaviate: %v\n", live)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func instantiateSchema(client *weaviate.Client, vectorizer string) error {
	classObj := &models.Class{
		Class:      "Question",
		Vectorizer: vectorizer,
	}

	if vectorizer != "none" {
		classObj.ModuleConfig = map[string]interface{}{
			vectorizer: map[string]interface{}{},
		}
	}

	err := client.Schema().ClassCreator().WithClass(classObj).Do(context.Background())
	if err != nil {
		return fmt.Errorf("failed to instantiate %s schema: %w", vectorizer, err)
	}
	return nil
}
