# Vector Databases

Experimenting ...

## Weaviate

- <https://weaviate.io/developers/weaviate/installation/docker-compose>
- <https://weaviate.io/developers/weaviate/quickstart>
- <https://platform.openai.com/docs/guides/embeddings>

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

## Chromem-go

An embedded vector database.

- <https://pkg.go.dev/github.com/philippgille/chromem-go>

```sh
source .env
go run cmd/chromem-go/main.go "topics on kubernetes security
```

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
https://technicalwriting.dev/rss.xml
https://sive.rs/en.atom

- Pocket Exports https://getpocket.com/export/
-->

<!--
TODO:
- RSS Search
  - Serverless deployment
    - https://cloud.google.com/kubernetes-engine/pricing#cluster_management_fee_and_free_tier
    - ServiceA = handling the query from the user
    - ServiceB = creating the vector database. NOTE. This won't work. Every query to the vector database will create a new vector database.
  - Only insert RSS feeds if it is not already in the DB. And if it has not been updated recently.
  - MAX_CONTENT_LENGTH should be defined in tokens not chars. Create multiple chunks for this blog post. https://github.com/openai/tiktoken
  - only return responses if they meet a certain similarity score?
  - grpc instead of http
  - Two APIs. One for updating the RSS feeds. One for searching the RSS feeds.
  - https://weaviate.io/developers/weaviate/configuration/backups
- Chromem-go
  - Insert an RSS feed item if it has been updated.
  - Only return responses meeting certain similarity score.
  - Function to return all RSS feeds.
  - Function to return all RSS feeds sorted by date.
  - Deploy to Google Cloud Run
  - Ability for users to create accounts. Ability for users to create their own list of RSS feeds.
- Ideas for using a VectorDB?
    - Storage & search. Open source codebases.
- Write a basic VectorDB. Insert embedding. Search for embedding.
- Creating your own Embeddings.
  - https://github.com/ynqa/wego
  - https://cybernetist.com/2024/01/07/fun-with-embeddings/
-->

<!--
DONE (new to old):
- Chromem-go RSS Search
  - Tokenizing & chunking
    - https://pkg.go.dev/github.com/pkoukk/tiktoken-go
    - https://pkg.go.dev/github.com/tmc/langchaingo/textsplitter
  - Only insert RSS feeds if it is not already in the DB.
  - Persist the DB to cloud storage. Load from cloud storage.
  - Persist the DB to disk. Load to disk.
- RSS Search
  - Add both "Item.Description" and "Item.Content" into the embedding
  - Embedding maxinput=8191 https://platform.openai.com/docs/guides/embeddings/embedding-models
  - Parsing RSS feed timestamps into RFC3339 format
  - Parameterize the RSS feeds
  - Generate a UUID for each RSS feed item?
  - Ensure batch import does not perform duplicate inserts.
- Insert & search for embedding using Weaviate Contextionary vectorizer.
- Deploy with Contextionary, OpenAI, standalone
-->