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

```sh
source .env
go run cmd/main/main.go
```

<!--
REFERENCES:
- https://weaviate.io/developers/weaviate/installation/docker-compose
- https://weaviate.io/developers/weaviate/quickstart
- https://platform.openai.com/docs/guides/embeddings
-->

<!--
RSS Feeds I follow:
https://thomasvn.dev/feed/
https://jvns.ca/atom.xml
https://golangweekly.com/rss/
https://blog.pragmaticengineer.com/feed/
https://rss.beehiiv.com/feeds/gQxaV1KHkQ.xml
https://world.hey.com/dhh/feed.atom
https://blog.kubecost.com/feed.xml
https://kubernetes.io/feed.xml

- Pocket Exports https://getpocket.com/export/
-->

<!-- 
2024/10/13 19:13:15 Configuring Weaviate connection ...
2024/10/13 19:13:15 Class RssFeeds already exists, skipping schema creation ...
2024/10/13 19:13:18 Batch inserting 159 objects ...
2024/10/13 19:13:38 failed to batch write objects: {
  "error": [
    {
      "message": "connection to: OpenAI API failed with status: 400 error: This model's maximum context length is 8192 tokens, however you requested 10905 tokens (10905 in your prompt; 0 for the completion). Please reduce your prompt; or completion length."
    }
  ]
}
exit status 1
-->

<!-- 
TODO:
- RSS Search
  - Embedding maxinput=8191 https://platform.openai.com/docs/guides/embeddings/embedding-models
  - only return responses if they meet a certain similarity score?
  - grpc instead of http
- Ideas for using a VectorDB?
    - Storage & search. Open source codebases.
- Write a basic VectorDB. Insert embedding. Search for embedding.
- Creating your own Embeddings.
  - https://github.com/ynqa/wego
  - https://cybernetist.com/2024/01/07/fun-with-embeddings/
-->

<!--
DONE (new to old):
- RSS Search
  - Parsing RSS feed timestamps into RFC3339 format
  - Parameterize the RSS feeds
  - Generate a UUID for each RSS feed item?
  - Ensure batch import does not perform duplicate inserts.
- Insert & search for embedding using Weaviate Contextionary vectorizer.
- Deploy with Contextionary, OpenAI, standalone
-->