# Zerops MCP Server v3

A Model Context Protocol (MCP) server for managing Zerops platform resources. This server provides 40 comprehensive tools for complete project lifecycle management including authentication, project management, service orchestration, deployment, configuration, and intelligent knowledge assistance.

## Features

- **40 Comprehensive Tools** across 8 categories
- **Direct API Integration** with Zerops platform
- **VPN Management** via zcli wrapper
- **Template System** for 6+ frameworks
- **Workflow Automation** for common tasks
- **Rich Error Messages** with actionable resolutions
- **Intelligent Knowledge System** with comprehensive platform documentation
- **Framework Patterns** for Laravel, Django, Next.js, Express, FastAPI
- **Configuration Validation** and dependency resolution

## Installation

### Prerequisites

- zcli (Zerops CLI) for deployment operations
- Valid Zerops API key

### Build from Source

Requires Go 1.21 or higher:

```bash
git clone https://github.com/krls2020/zerops-mcp.git
cd zerops-mcp
make build
```

### Environment Setup

## Quick Start

1. **Set up your API key:**
   ```bash   
   export ZEROPS_API_KEY="your-api-key"
   claude mcp add zerops -s user [path-to-mcp-server-folder]/mcp-server
   ```

### Project Structure

```
zerops-mcp/
├── internal/
│   ├── api/            # Zerops API client
│   ├── tools/          # MCP tool implementations
│   ├── zcli/           # zcli wrapper
│   ├── templates/      # Configuration templates
│   └── config/         # Server configuration
├── test/               # Integration tests
└── docs/               # Documentation
```

## Contributing

1. Fork the repository
2. Create your feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

[License information here]

## Support

For issues and questions:
- GitHub Issues: [github.com/zeropsio/zerops-mcp-v3/issues]
- Zerops Documentation: [docs.zerops.io]