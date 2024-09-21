package main

import (
	"context"
	"fmt"
	"os"

	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/auth"
	"github.com/weaviate/weaviate/entities/models"
)

var (
	WEAVIATE_URL     = os.Getenv("WEAVIATE_URL")     // "localhost:8080"
	WEAVIATE_API_KEY = os.Getenv("WEAVIATE_API_KEY") // ""
	OPENAI_API_KEY   = os.Getenv("OPENAI_API_KEY")   // ""
)

func main() {
	// Create connection
	cfg := weaviate.Config{
		Host:       WEAVIATE_URL,
		Scheme:     "http",
		AuthConfig: auth.ApiKey{Value: WEAVIATE_API_KEY},
		Headers: map[string]string{
			"X-OpenAI-Api-Key": OPENAI_API_KEY,
		},
	}
	client, err := weaviate.NewClient(cfg)
	if err != nil {
		panic(err)
	}

	// Check the connection
	live, err := client.Misc().LiveChecker().Do(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v", live)

	// Create schema
	classObj := &models.Class{
		Class:      "Question",
		Vectorizer: "text2vec-contextionary", // "none", "text2vec-contextionary", "text2vec-openai"
		ModuleConfig: map[string]interface{}{
			"text2vec-contextionary": map[string]interface{}{},
		},
	}
	err = client.Schema().ClassCreator().WithClass(classObj).Do(context.Background())
	if err != nil {
		panic(err)
	}
}
