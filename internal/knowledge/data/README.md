# Zerops Service Knowledge Base

This directory contains comprehensive JSON knowledge files for all Zerops service types. These files serve as the source of truth for service configurations, best practices, and examples.

## Structure

```
services/
├── index.json          # Master index of all services
├── nodejs.json         # Node.js runtime
├── php.json           # PHP runtime  
├── python.json        # Python runtime
├── go.json            # Go runtime
├── dotnet.json        # .NET runtime
├── rust.json          # Rust runtime
├── java.json          # Java runtime
├── ruby.json          # Ruby runtime
├── container.json     # Custom containers
├── postgresql.json    # PostgreSQL database
├── mariadb.json       # MariaDB database
├── mongodb.json       # MongoDB database
├── valkey.json        # Valkey (Redis fork)
├── keydb.json         # KeyDB (Redis fork)
├── qdrant.json        # Qdrant vector database
├── couchbase.json     # Couchbase database
├── elasticsearch.json # Elasticsearch
├── meilisearch.json   # Meilisearch
├── typesense.json     # Typesense
├── nats.json          # NATS messaging
├── kafka.json         # Apache Kafka
├── rabbitmq.json      # RabbitMQ
└── static.json        # Static file hosting
```

## Service File Format

Each service JSON file contains:

- **type**: Service identifier
- **category**: Service category (runtime, database, search, messaging, web)
- **description**: Brief description
- **versions**: Available versions with status
- **modes**: HA/NON_HA support
- **ports**: Port configuration
- **configuration**: Build and runtime settings
- **envVariables**: System and auto-generated variables
- **bestPractices**: Recommended practices
- **examples**: Complete configuration examples
- **useCases**: Common use cases

## Usage

These files can be used by:
1. MCP tools to provide accurate service information
2. Documentation generation
3. Configuration validation
4. Template generation

## Adding New Services

When adding a new service:
1. Create a new JSON file following the existing format
2. Update `index.json` to include the new service
3. Include all required fields
4. Add comprehensive examples
5. Document best practices

## Updating Services

When updating service information:
1. Update the specific service JSON file
2. Verify version information is current
3. Update examples if needed
4. Document any breaking changes