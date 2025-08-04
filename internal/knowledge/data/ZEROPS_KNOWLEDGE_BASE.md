# Zerops Platform Knowledge Base

## Table of Contents
1. [Introduction](#introduction)
2. [Features](#features)
3. [Service Types](#service-types)
4. [Core Concepts](#core-concepts)
5. [Configuration](#configuration)
6. [Deployment](#deployment)
7. [Networking](#networking)
8. [Environment Variables](#environment-variables)
9. [Build Pipeline](#build-pipeline)
10. [Automatic Scaling & High Availability](#automatic-scaling--high-availability)
11. [Backup](#backup)
12. [CDN](#cdn)
13. [Command Line Tools](#command-line-tools)
14. [Access & Connectivity](#access--connectivity)
15. [API & Integrations](#api--integrations)
16. [Observability](#observability)
17. [Zerops.yml Specification](#zeropsyml-specification)
18. [Laravel-Specific Optimizations](#laravel-specific-optimizations)
19. [Best Practices](#best-practices)
20. [Common Issues](#common-issues)

## Introduction

Zerops is a developer-first cloud platform that simplifies application deployment and infrastructure management. It provides a fully managed environment with automatic scaling, high availability, and built-in DevOps practices, allowing developers to focus on code rather than infrastructure.

### Key Principles
- **Zero DevOps**: Infrastructure management is fully automated
- **Developer-Focused**: Optimized workflows for modern development
- **Language Agnostic**: Supports all major programming languages and frameworks
- **Production-Ready**: Enterprise-grade reliability and performance
- **Cost-Efficient**: Pay only for resources actually used

### Platform Philosophy
Zerops eliminates the complexity of traditional cloud platforms by providing:
- Pre-configured environments for popular frameworks
- Automatic security updates and patches
- Built-in monitoring and alerting
- Seamless scaling without configuration
- Integrated development tools

## Features

### Core Platform Features
- **Automatic Scaling**: Horizontal and vertical scaling based on load
- **High Availability**: Built-in redundancy and failover
- **Zero-Downtime Deployments**: Seamless application updates
- **Private Networking**: Isolated VXLAN networks per project
- **Managed Databases**: Automated backups, updates, and optimization
- **Built-in Monitoring**: Real-time metrics and logging
- **Global CDN**: Integrated content delivery network
- **SSL Management**: Automatic certificate provisioning and renewal

### Developer Experience
- **Git-Based Deployments**: Direct integration with GitHub/GitLab
- **Local Development**: VPN access to production environment
- **Multiple Environments**: Easy staging/production separation
- **Team Collaboration**: Role-based access control
- **API Access**: Full programmatic control
- **CLI Tools**: Command-line interface for automation

### Security Features
- **Network Isolation**: Each project in private network
- **Automatic Updates**: Security patches applied automatically
- **Encrypted Storage**: All data encrypted at rest
- **DDoS Protection**: Built-in protection against attacks
- **Firewall Rules**: Granular access control
- **Secret Management**: Secure handling of sensitive data

### Performance Optimization
- **Smart Caching**: Automatic caching strategies
- **Load Balancing**: Intelligent request distribution
- **Resource Optimization**: Efficient resource utilization
- **Global Infrastructure**: Multiple regions available
- **Edge Computing**: CDN with compute capabilities

## Service Types

### Runtime Services

#### Node.js
- **Versions**: `nodejs@22`, `nodejs@20`, `nodejs@18`
- **Key Features**:
  - NPM/Yarn/PNPM support
  - Built-in PM2 for process management
  - Automatic node_modules caching
  - Support for monorepos

#### PHP
- **Apache Versions**: `php-apache@8.4`, `php-apache@8.3`, `php-apache@8.1`
- **Nginx Versions**: `php-nginx@8.4`, `php-nginx@8.3`, `php-nginx@8.1`
- **Key Features**:
  - Composer support
  - Laravel optimizations
  - OPCache enabled
  - Document root: `/var/www/public`

#### Python
- **Versions**: `python@3.12`, `python@3.11`
- **Key Features**:
  - Pip/Poetry support
  - Virtual environment auto-creation
  - WSGI/ASGI server support
  - Requirements.txt auto-detection

#### Go
- **Versions**: `go@1.22`, `go@1`, `golang@1`
- **Key Features**:
  - Module support
  - Vendor directory caching
  - Build optimization
  - Static binary compilation

#### .NET
- **Versions**: `dotnet@9`, `dotnet@8`, `dotnet@7`, `dotnet@6`
- **Key Features**:
  - ASP.NET Core support
  - NuGet package management
  - Runtime optimization
  - Multi-project solutions

#### Rust
- **Versions**: `rust@1`, `rust@1.78`, `rust@nightly`
- **Key Features**:
  - Cargo support
  - Target directory caching
  - Release builds by default
  - Cross-compilation support

#### Java
- **Versions**: `java@21`, `java@17`
- **Key Features**:
  - Maven/Gradle support
  - JAR/WAR deployment
  - JVM optimization
  - Spring Boot ready

#### Other Runtimes
- **Deno**: `deno@2.0.0`, `deno@1.45.5`
- **Bun**: `bun@1.2.2`, `bun@1.1.34`, `bun@nightly`, `bun@canary`
- **Elixir**: `elixir@1`, `elixir@1.16`
- **Gleam**: `gleam@1`, `gleam@1.5`

### Static Services
- **Nginx**: `nginx@1.22` - For static websites
- **Static**: `static`, `static@1.0` - Simple file serving

### Container Services
- **Alpine**: `alpine@3.20`, `alpine@3.19`, `alpine@3.18`, `alpine@3.17`
- **Ubuntu**: `ubuntu@24.04`, `ubuntu@22.04`
- **Docker**: `docker@26.1.5` - Docker-in-Docker support

### Database Services

#### PostgreSQL
- **Versions**: `postgresql@17`, `postgresql@16`, `postgresql@14`
- **Auto-generated Variables**:
  - `hostname`, `port`, `user`, `password`, `dbName`
- **Features**: Full-text search, JSON support, extensions

#### MariaDB
- **Version**: `mariadb@10.6` (ONLY 10.6 supported!)
- **Auto-generated Variables**:
  - `hostname`, `port`, `user`, `password`, `dbName`
- **Note**: Version 11 NOT supported

#### Cache Stores
- **Valkey**: `valkey@7.2` - Redis alternative
- **KeyDB**: `keydb@6` - Multi-threaded Redis alternative
- **Note**: Redis is NOT available, use Valkey or KeyDB

#### Vector Database
- **Qdrant**: `qdrant@1.12`, `qdrant@1.10`
- **Features**: Vector similarity search, collections, REST API

### Search Engines
- **Elasticsearch**: `elasticsearch@8.16`
- **Meilisearch**: `meilisearch@1.10`
- **Typesense**: `typesense@27.1`

### Message Brokers
- **NATS**: `nats@2.10` - Lightweight messaging
- **Kafka**: `kafka@3.8` - Distributed streaming

### Storage Services
- **Object Storage**: `object-storage` - S3-compatible
- **Shared Storage**: `shared-storage` - NFS-like persistent storage

### Analytics
- **ClickHouse**: `clickhouse@25.3` - Column-oriented database

## Core Concepts

### Project Structure
```
Project (isolated environment)
├── Service 1 (app)
│   └── Container(s)
├── Service 2 (database)
│   └── Container(s)
└── Service 3 (cache)
    └── Container(s)
```

### Service Modes
- **HA (High Availability)**: Multiple containers, auto-scaling
- **NON_HA**: Single container, suitable for dev/small projects

### Project Core Types
- **Lightweight Core**: Single-container control plane
- **Serious Core**: Multi-container HA infrastructure

## Configuration

### Service YAML Structure
```yaml
services:
  - hostname: app
    type: nodejs@20
    mode: HA
    minContainers: 2
    maxContainers: 5
    ports:
      - port: 3000
        httpSupport: true
    enableSubdomainAccess: true
    envVariables:
      NODE_ENV: production
      DB_HOST: ${db_hostname}
```

### Deployment Configuration (zerops.yml)
```yaml
zerops:
  - setup: app
    build:
      base: nodejs@20
      buildCommands:
        - npm ci
        - npm run build
      deployFiles:
        - dist
        - package.json
        - node_modules
    run:
      base: nodejs@20
      ports:
        - port: 3000
          httpSupport: true
      start: npm start
```

## Deployment

### Methods
1. **zcli push**: Local deployment via CLI
2. **GitHub/GitLab Integration**: Automatic deployments
3. **buildFromGit**: One-time build from repository (recipes only)

### Requirements
- Git repository with at least one commit
- `zerops.yml` in repository root
- VPN connection for local deployments
- Files deployed to `/var/www`

### Zero-Downtime Process
1. New containers start with new version
2. Readiness checks performed
3. Traffic gradually shifts to new containers
4. Old containers gracefully shutdown

## Networking

### Public Access Options

#### Zerops Subdomain
- Format: `https://app-[project].zerops.app`
- Auto-SSL management
- 50MB upload limit
- For development/testing

#### Custom Domains
- Production-ready
- L7 load balancer
- Let's Encrypt or custom SSL
- IPv4 (shared/dedicated) + IPv6

#### Direct Port Access
- Ports 10-65435 (except 80, 443)
- For non-HTTP protocols
- Optional firewall rules

### Internal Networking
- Private VXLAN network per project
- Service-to-service via hostnames
- No external access by default

## Environment Variables

### Types
- **User-defined**: Set in YAML or GUI
- **Secret**: Hidden in GUI, for sensitive data
- **System-generated**: Auto-created by Zerops

### Cross-Service References
```yaml
# Default (with isolation)
envVariables:
  DB_HOST: ${db_hostname}
  DB_PASS: ${db_password}

# Without isolation
envVariableIsolation: false
# Variables auto-available as servicename_variablename
```

### Special Variables
- `${zeropsSubdomain}`: Auto-generated subdomain URL
- `${BUILD_*}`: Access build vars in runtime
- `${RUNTIME_*}`: Access runtime vars in build

## Build Pipeline

### Build Steps
1. Create temporary build container
2. Install build base environment
3. Clone/download source code
4. Run `build.prepareCommands`
5. Execute `build.buildCommands`
6. Package `build.deployFiles`
7. Cache specified directories

### Build Configuration
```yaml
build:
  base: nodejs@20
  os: alpine  # or ubuntu
  prepareCommands:
    - npm config set registry https://custom.registry
  buildCommands:
    - npm ci
    - npm run build
  deployFiles:
    - dist
    - package.json
  cache:
    - node_modules
    - .npm
```

### Runtime Preparation
```yaml
build:
  addToRunPrepare:
    - package.json
    - package-lock.json
run:
  prepareCommands:
    - npm ci --production
```

## Best Practices

### Service Naming
- ✅ **Good**: `app`, `db01`, `cache1`
- ❌ **Bad**: `my-app`, `web_server`, `MyApp`
- **Rule**: Only lowercase letters and numbers

### Database Selection
- **PostgreSQL**: Recommended for most applications
- **MariaDB**: Use only version 10.6
- **Cache**: Use Valkey or KeyDB (not Redis)

### Framework-Specific

#### Laravel (PHP)
```yaml
services:
  - hostname: app
    type: php-apache@8.3
    envVariables:
      APP_ENV: production
      DB_CONNECTION: pgsql
      DB_HOST: ${db_hostname}
      DB_PASSWORD: ${db_password}
```

#### Django (Python)
```yaml
run:
  prepareCommands:
    - python -m pip install -r requirements.txt
  start: gunicorn myapp.wsgi:application
```

#### Next.js (Node.js)
```yaml
build:
  buildCommands:
    - npm ci
    - npm run build
run:
  start: npm start
  envVariables:
    NODE_ENV: production
```

### Performance Optimization
1. Use caching in build configuration
2. Minimize deployFiles size
3. Use prepareCommands for runtime optimization
4. Enable auto-scaling for production

## Common Issues

### Build Failures
- **"Could not open requirements file"**: Add to `addToRunPrepare`
- **"exit status 128"**: Git needs at least one commit
- **"websocket: bad handshake"**: Log streaming issue, not failure

### Service Issues
- **"Service name invalid"**: Use only lowercase letters/numbers
- **"No action allowed, project will be deleted"**: Wait for initialization
- **"Service stack is not http"**: Add `httpSupport: true` to ports

### VPN Issues
- **"VPN connection failed"**: Check zcli installation and sudo access
- **"Project not found"**: Verify project ID with project_list

### Deployment Issues
- **Files not found**: Ensure proper `deployFiles` configuration
- **Start command fails**: Check file paths relative to `/var/www`
- **Environment variables undefined**: Check service references

## Service Configuration Examples

### Full Stack Application
```yaml
services:
  - hostname: api
    type: nodejs@20
    ports:
      - port: 3000
        httpSupport: true
    envVariables:
      DB_HOST: ${db_hostname}
      DB_PASSWORD: ${db_password}
      REDIS_HOST: ${cache_hostname}
  
  - hostname: db
    type: postgresql@16
    mode: HA
  
  - hostname: cache
    type: valkey@7.2
```

### Microservices Pattern
```yaml
services:
  - hostname: gateway
    type: nodejs@20
    ports:
      - port: 8080
        httpSupport: true
    envVariables:
      AUTH_SERVICE: http://auth:3000
      USER_SERVICE: http://users:3000
  
  - hostname: auth
    type: golang@1
    ports:
      - port: 3000
  
  - hostname: users
    type: python@3.12
    ports:
      - port: 3000
```

### Static Website with CDN
```yaml
services:
  - hostname: web
    type: static
    enableSubdomainAccess: true
    ports:
      - port: 80
        httpSupport: true
```

## Routing and Load Balancing

### HTTP Routing
- Automatic load balancing across containers
- Health check-based routing
- Session affinity available
- WebSocket support

### Path-Based Routing
```yaml
routes:
  - path: /api
    service: api
  - path: /admin
    service: admin
  - path: /
    service: frontend
```

### Custom Headers
- Add security headers
- CORS configuration
- Cache control
- Custom response headers

## Security Features

### Network Security
- Isolated project networks
- Firewall rules for public ports
- DDoS protection
- SSL/TLS termination

### Application Security
- Secret environment variables
- Secure build process
- Container isolation
- Regular security updates

### Compliance
- GDPR compliant infrastructure
- Data residency options
- Audit logging
- Backup encryption

## Monitoring and Logs

### Log Types
- **Build logs**: Compilation and packaging
- **Runtime logs**: Application output
- **System logs**: Container events

### Log Access
- Real-time streaming
- Historical search
- Log export options
- Integration with external tools

### Metrics
- CPU and memory usage
- Network traffic
- Response times
- Error rates

## Advanced Features

### Cron Jobs
```yaml
run:
  crontab:
    - spec: "0 2 * * *"
      command: "php artisan backup:daily"
```

### Health Checks
```yaml
run:
  healthCheck:
    httpGet:
      path: /health
      port: 3000
    intervalSeconds: 30
    timeoutSeconds: 5
```

### Custom Runtime Images
```yaml
run:
  base: ubuntu@24.04
  prepareCommands:
    - apt-get update
    - apt-get install -y custom-package
```

### Multi-Stage Builds
```yaml
build:
  base: golang@1
  buildCommands:
    - go build -o app
  deployFiles:
    - app
run:
  base: alpine@3.20
  start: ./app
```

## Automatic Scaling & High Availability

### Horizontal Scaling
- **Container Scaling**: Automatically add/remove containers based on load
- **Configuration**:
  ```yaml
  services:
    - hostname: app
      type: nodejs@20
      mode: HA
      minContainers: 2
      maxContainers: 10
      scalingRules:
        cpu: 80  # Scale up at 80% CPU
        memory: 85  # Scale up at 85% memory
  ```

### Vertical Scaling
- **Resource Adjustment**: CPU and memory automatically adjusted
- **No Downtime**: Resources adjusted without service interruption
- **Limits**: Set maximum resources per container

### High Availability Features
- **Multi-Container Mode**: Minimum 2 containers in HA mode
- **Load Distribution**: Automatic request distribution
- **Health Monitoring**: Continuous health checks
- **Automatic Recovery**: Failed containers automatically replaced
- **Zero-Downtime Updates**: Rolling deployments

## Backup

### Automatic Backups
- **Databases**: Daily automatic backups for all database services
- **Retention**: 30 days by default
- **Point-in-Time Recovery**: For PostgreSQL and MariaDB
- **Encryption**: All backups encrypted at rest

### Manual Backups
- **On-Demand**: Create backups anytime via GUI/API
- **Export**: Download backups for offline storage
- **Cross-Region**: Copy backups to different regions

### Restoration
- **One-Click Restore**: Restore from any backup point
- **New Service**: Restore to a new database service
- **Selective Restore**: Choose specific databases/tables

## CDN

### Integrated CDN
- **Global Edge Network**: 100+ PoPs worldwide
- **Automatic Caching**: Static assets cached automatically
- **SSL/TLS**: Full HTTPS support
- **Compression**: Automatic Gzip/Brotli compression

### Configuration
```yaml
services:
  - hostname: web
    type: static
    cdn:
      enabled: true
      cacheControl: "public, max-age=31536000"
      compressionLevel: 9
      customHeaders:
        X-Frame-Options: "DENY"
        X-Content-Type-Options: "nosniff"
```

### Cache Management
- **Purge Cache**: Clear CDN cache on demand
- **Cache Rules**: Custom caching per file type
- **Origin Shield**: Reduce origin server load

## Command Line Tools

### zCLI (Zerops CLI)
Primary command-line tool for Zerops platform interaction.

#### Installation
```bash
# macOS/Linux
curl -L https://zerops.io/zcli/install.sh | sh

# Windows
iwr -useb https://zerops.io/zcli/install.ps1 | iex
```

#### Key Commands
```bash
# Authentication
zcli login

# Project Management
zcli project list
zcli project create <name>

# Service Management
zcli service list --projectId <id>
zcli service log <serviceName> --projectId <id>

# Deployment
zcli push --projectId <id>

# VPN Management
zcli vpn up --projectId <id>
zcli vpn down
zcli vpn status
```

#### Configuration File (.zerops.yml)
```yaml
projectId: your-project-id
workingDir: ./
configPath: ./zerops.yml
```

### zsc (Zerops Service Client)
Lightweight tool for service interaction and debugging.

#### Features
- **Direct Service Access**: Connect to service containers
- **File Transfer**: Upload/download files
- **Process Management**: View and manage processes
- **Real-time Logs**: Stream service logs

#### Usage
```bash
# Connect to service
zsc connect <serviceName>

# Transfer files
zsc upload <localFile> <serviceName>:<remotePath>
zsc download <serviceName>:<remotePath> <localFile>

# View processes
zsc ps <serviceName>

# Stream logs
zsc logs <serviceName> -f
```

## Access & Connectivity

### SSH Access
Direct SSH access to service containers for debugging and maintenance.

#### Enabling SSH
```yaml
services:
  - hostname: app
    type: nodejs@20
    sshAccess: true
    authorizedKeys:
      - "ssh-rsa AAAAB3NzaC1yc2EA... user@host"
```

#### Connection
```bash
# Via zcli
zcli ssh <serviceName> --projectId <id>

# Direct SSH (requires VPN)
ssh zerops@<service-hostname>
```

### VPN Access
Secure connection to project's private network.

#### Features
- **Full Network Access**: Access all project services
- **Development Integration**: Test locally with production services
- **Secure Tunnel**: WireGuard-based VPN
- **Multi-User**: Multiple team members can connect

#### Setup
```bash
# Connect to project VPN
sudo zcli vpn up --projectId <id>

# Check connection status
zcli vpn status

# Disconnect
sudo zcli vpn down
```

### Firewall Rules
Control access to services with granular firewall rules.

```yaml
services:
  - hostname: api
    type: nodejs@20
    firewall:
      - source: 0.0.0.0/0
        port: 443
        protocol: tcp
      - source: 10.0.0.0/8
        port: 22
        protocol: tcp
```

### SMTP Configuration
Send emails from your applications.

#### Outbound SMTP
- **Port 25**: Blocked by default (anti-spam)
- **Port 587/465**: Available for authenticated SMTP
- **Rate Limiting**: 100 emails/hour by default

#### Configuration
```yaml
envVariables:
  SMTP_HOST: smtp.sendgrid.net
  SMTP_PORT: 587
  SMTP_USER: ${smtp_user}
  SMTP_PASS: ${smtp_password}
```

## API & Integrations

### Zerops API
RESTful API for complete platform control.

#### Authentication
```bash
curl -X POST https://api.app-prg1.zerops.io/v1/auth/token \
  -H "Content-Type: application/json" \
  -d '{"apiKey": "your-api-key"}'
```

#### Common Operations
```bash
# List projects
GET /v1/projects

# Create service
POST /v1/project/{projectId}/service

# Get service logs
GET /v1/project/{projectId}/service/{serviceId}/logs

# Trigger deployment
POST /v1/project/{projectId}/service/{serviceId}/deploy
```

### Import & Export

#### Import File Format
```yaml
# import.yml
project:
  name: my-project
  region: prg1
services:
  - hostname: app
    type: nodejs@20
    mode: HA
    # ... full service configuration
```

#### Import Command
```bash
zcli project import --file import.yml
```

#### Export Project
```bash
zcli project export --projectId <id> --output export.yml
```

### GitHub Integration

#### Setup
1. Install Zerops GitHub App
2. Configure in project settings
3. Select repository and branch
4. Automatic deployments on push

#### Configuration (.github/zerops.yml)
```yaml
deploy:
  - branch: main
    target: production
  - branch: develop
    target: staging
```

### GitLab Integration

#### Setup
1. Add GitLab webhook to project
2. Configure deployment token
3. Set branch rules
4. Automatic deployments on push

#### .gitlab-ci.yml Example
```yaml
deploy:
  stage: deploy
  script:
    - zcli push --projectId $ZEROPS_PROJECT_ID
  only:
    - main
```

## Observability

### Logging

#### Log Types
- **Application Logs**: stdout/stderr from containers
- **Build Logs**: Build process output
- **System Logs**: Platform events
- **Access Logs**: HTTP request logs

#### Log Management
```bash
# View logs
zcli service log <serviceName> --projectId <id>

# Follow logs
zcli service log <serviceName> --projectId <id> -f

# Filter logs
zcli service log <serviceName> --projectId <id> --filter "ERROR"

# Download logs
zcli service log <serviceName> --projectId <id> --output logs.txt
```

### Log Forwarding

#### Supported Destinations
- **Elasticsearch**: Direct integration
- **Datadog**: API key configuration
- **New Relic**: License key required
- **Custom Webhook**: POST to any endpoint

#### Configuration
```yaml
services:
  - hostname: app
    type: nodejs@20
    logForwarding:
      - type: elasticsearch
        url: https://logs.example.com
        index: zerops-logs
        apiKey: ${log_api_key}
      - type: datadog
        apiKey: ${datadog_api_key}
        tags:
          - env:production
          - app:myapp
```

#### Syslog Forwarding
```yaml
logForwarding:
  - type: syslog
    host: syslog.example.com
    port: 514
    protocol: tcp
    format: rfc5424
```

## Zerops.yml Specification

### Complete Reference

#### Root Structure
```yaml
zerops:
  - setup: <service-name>
    build: <build-config>
    run: <runtime-config>
    deploy: <deploy-config>
```

#### Build Configuration
```yaml
build:
  base: <technology>@<version>
  os: alpine|ubuntu
  prepareCommands: []
  buildCommands: []
  deployFiles: []
  addToRunPrepare: []
  cache: []
  envVariables: {}
```

#### Runtime Configuration
```yaml
run:
  base: <technology>@<version>
  os: alpine|ubuntu
  prepareCommands: []
  initCommands: []
  start: <start-command>
  ports:
    - port: <number>
      httpSupport: true|false
      protocol: tcp|udp
  healthCheck: {}
  readinessCheck: {}
  livenessCheck: {}
  crontab: []
  envVariables: {}
```

### Base List
Complete list of available base images:
- **Node.js**: nodejs@22, nodejs@20, nodejs@18, nodejs@16
- **PHP**: php@8.4, php@8.3, php@8.2, php@8.1, php@8.0, php@7.4
- **Python**: python@3.12, python@3.11, python@3.10, python@3.9
- **Go**: go@1.22, go@1.21, go@1.20, go@1.19
- **Ruby**: ruby@3.3, ruby@3.2, ruby@3.1, ruby@3.0
- **Java**: java@21, java@17, java@11, java@8
- **Rust**: rust@1, rust@nightly
- **.NET**: dotnet@8, dotnet@7, dotnet@6
- **OS**: alpine@3.20, ubuntu@24.04, ubuntu@22.04

### Cron Configuration
```yaml
run:
  crontab:
    - spec: "0 */6 * * *"  # Every 6 hours
      command: "php artisan cache:clear"
    - spec: "0 2 * * *"    # Daily at 2 AM
      command: "python manage.py clearsessions"
    - spec: "@hourly"      # Every hour
      command: "node scripts/cleanup.js"
```

### References & Variables
```yaml
# Service references
${servicename_variable}  # Reference another service's variable
${servicename}          # Reference service hostname

# Special variables
${zeropsSubdomain}      # Auto-generated subdomain
${projectId}            # Current project ID
${serviceId}            # Current service ID
${containerId}          # Current container ID

# Build/Runtime context
${BUILD_COMMIT_HASH}    # Git commit hash
${BUILD_TIMESTAMP}      # Build timestamp
${RUNTIME_PORT}         # Primary port number
```

## Laravel-Specific Optimizations

### Recommended Configuration
```yaml
services:
  - hostname: app
    type: php-apache@8.3
    mode: HA
    minContainers: 2
    enableSubdomainAccess: true
    envVariables:
      APP_ENV: production
      APP_DEBUG: false
      LOG_CHANNEL: stack
      DB_CONNECTION: pgsql
      DB_HOST: ${db_hostname}
      DB_PASSWORD: ${db_password}
      CACHE_DRIVER: redis
      REDIS_HOST: ${cache_hostname}
      SESSION_DRIVER: redis
      QUEUE_CONNECTION: redis
```

### Deployment Configuration
```yaml
zerops:
  - setup: app
    build:
      base: php@8.3
      buildCommands:
        - composer install --optimize-autoloader --no-dev
        - npm ci && npm run production
        - php artisan config:cache
        - php artisan route:cache
        - php artisan view:cache
        - php artisan event:cache
      deployFiles:
        - .
        - "!node_modules"
        - "!tests"
        - "!.git"
    run:
      base: php-apache@8.3
      documentRoot: public
      initCommands:
        - php artisan migrate --force --isolated
        - php artisan storage:link
        - php artisan queue:restart
        - php artisan horizon:terminate
```

### Queue Workers
```yaml
services:
  - hostname: queue
    type: php@8.3
    envVariables:
      # Same as app service
    run:
      start: php artisan queue:work --tries=3 --timeout=90
```

### Laravel Horizon
```yaml
services:
  - hostname: horizon
    type: php@8.3
    envVariables:
      # Same as app service
    run:
      start: php artisan horizon
```

This comprehensive knowledge base now covers all major aspects of the Zerops platform as outlined in the documentation structure.