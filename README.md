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
TODO:
- RSS Search
  - Parameterize the RSS feeds.
  - Ensure we are not doing unnecessary duplicate inserts.
  - Generate a UUID for each RSS feed item? https://weaviate.io/developers/weaviate/manage-data/import#specify-an-id-value
- Ideas for using a VectorDB?
    - Storage & search. Open source codebases.
- Write a basic VectorDB. Insert embedding. Search for embedding.
- Creating your own Embeddings.
  - https://github.com/ynqa/wego
  - https://cybernetist.com/2024/01/07/fun-with-embeddings/
-->

<!--
DONE (new to old):
- Insert & search for embedding using Weaviate Contextionary vectorizer.
- Deploy with Contextionary, OpenAI, standalone
-->