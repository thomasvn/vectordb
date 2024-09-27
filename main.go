package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
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
		log.Fatalf("Failed to create Weaviate client: %v", err)
	}

	if err := instantiateSchema(client, VECTORIZER); err != nil {
		log.Fatalf("Failed to instantiate schema: %v", err)
	}

	if err := insertData(client); err != nil {
		log.Fatalf("Failed to insert data: %v", err)
	}

	if err := semanticSearch(client); err != nil {
		log.Fatalf("Failed to perform semantic search: %v", err)
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
	className := "Question"

	exists, err := client.Schema().ClassExistenceChecker().WithClassName(className).Do(context.Background())
	if err != nil {
		return fmt.Errorf("failed to check if class exists: %w", err)
	}
	if exists {
		fmt.Println("Class 'Question' already exists, skipping schema creation")
		return nil
	}

	classObj := &models.Class{
		Class:      className,
		Vectorizer: vectorizer,
	}
	if vectorizer != "none" {
		classObj.ModuleConfig = map[string]interface{}{
			vectorizer: map[string]interface{}{},
		}
	}

	err = client.Schema().ClassCreator().WithClass(classObj).Do(context.Background())
	if err != nil {
		return fmt.Errorf("failed to instantiate %s schema: %w", vectorizer, err)
	}
	return nil
}

func insertData(client *weaviate.Client) error {
	// Retrieve the data
	data, err := http.DefaultClient.Get("https://raw.githubusercontent.com/weaviate-tutorials/quickstart/main/data/jeopardy_tiny.json")
	if err != nil {
		return fmt.Errorf("failed to retrieve data: %w", err)
	}
	defer data.Body.Close()

	// Decode the data
	var items []map[string]string
	if err := json.NewDecoder(data.Body).Decode(&items); err != nil {
		return fmt.Errorf("failed to decode data: %w", err)
	}

	// Convert items into a slice of models.Object
	objects := make([]*models.Object, len(items))
	for i := range items {
		objects[i] = &models.Object{
			Class: "Question",
			Properties: map[string]any{
				"category": items[i]["Category"],
				"question": items[i]["Question"],
				"answer":   items[i]["Answer"],
			},
		}
	}

	// Batch write items
	batchRes, err := client.Batch().ObjectsBatcher().WithObjects(objects...).Do(context.Background())
	if err != nil {
		return fmt.Errorf("failed to batch write objects: %w", err)
	}
	for _, res := range batchRes {
		if res.Result.Errors != nil {
			return fmt.Errorf("failed to batch write objects: %v", res.Result.Errors.Error)
		}
	}

	return nil
}

func semanticSearch(client *weaviate.Client) error {
	fields := []graphql.Field{
		{Name: "question"},
		{Name: "answer"},
		{Name: "category"},
	}

	nearText := client.GraphQL().
		NearTextArgBuilder().
		WithConcepts([]string{"beatles"})

	result, err := client.GraphQL().Get().
		WithClassName("Question").
		WithFields(fields...).
		WithNearText(nearText).
		WithLimit(10).
		Do(context.Background())
	if err != nil {
		return fmt.Errorf("failed to perform semantic search: %w", err)
	}

	jsonData, err := json.MarshalIndent(result.Data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal result data: %w", err)
	}
	fmt.Println("Semantic Search Results:")
	fmt.Println(string(jsonData))

	return nil
}
