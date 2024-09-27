# Vector Databases

Experimenting ...

## Weaviate

```sh
# Option1: Locally run VectorDB with no vectorizer module
docker run -p 8080:8080 -p 50051:50051 cr.weaviate.io/semitechnologies/weaviate:1.26.4

# Option2: Locally run VectorDB with text2vec-contextionary vectorizer
docker compose -f docker-compose.contextionary.yml up -d
docker compose -f docker-compose.contextionary.yml down

# Option3: Locally run VectorDB with text2vec-openai vectorizer
docker compose -f docker-compose.openai.yml up -d
docker compose -f docker-compose.openai.yml down
```

<!-- 
REFERENCES:
- https://weaviate.io/developers/weaviate/installation/docker-compose
- https://weaviate.io/developers/weaviate/quickstart
-->

<!-- 
TODO:
- Write a basic VectorDB. Insert embedding. Search for embedding.
- https://cybernetist.com/2024/01/07/fun-with-embeddings/
- https://github.com/ynqa/wego
-->

<!--
DONE (new to old):
- Insert & search for embedding using Weaviate Contextionary vectorizer.
- Deploy with Contextionary, OpenAI, standalone
-->