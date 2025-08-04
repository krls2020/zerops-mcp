# Deployment Scripts

This directory contains useful deployment scripts for the Zerops MCP Server.

## Scripts

### deploy_single_runtime.sh
Deploys a single runtime recipe to Zerops with basic functionality.

Usage:
```bash
./deploy_single_runtime.sh python
```

### deploy_all_runtimes_with_subdomains.sh
Deploys all 12 runtime recipes with subdomain access enabled.

Features:
- Creates projects with app + database services
- Deploys application code via zcli
- Automatically enables subdomain access
- Provides deployment summary

Usage:
```bash
./deploy_all_runtimes_with_subdomains.sh
```

### enable_subdomain_for_project.sh
Enables subdomain access for a specific service in a project.

Usage:
```bash
./enable_subdomain_for_project.sh <project_id> [service_name]
```

## Important Notes

1. **VPN Connection**: All deployment scripts require VPN connection via `zcli vpn up`
2. **Git Requirements**: Deployment directories need at least one git commit
3. **Subdomain Access**: Only works for services with HTTP/HTTPS ports
4. **API Key**: Scripts use the test API key from CLAUDE.md

## Testing

The enhanced deployment test is located at:
`test/integration/deploy_runtime_test.go`

This test:
- Creates projects with services
- Connects VPN
- Deploys code with proper git setup
- Enables subdomain access
- Handles common errors gracefully