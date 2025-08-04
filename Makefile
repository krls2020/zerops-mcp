BINARY=mcp-server
GO=/usr/local/go/bin/go
GOFLAGS=-v
TEST_TIMEOUT=30m

.PHONY: build test run clean lint fmt test-phase-% test-coverage test-integration test-runtime verify-runtime simulate-runtime

build:
	$(GO) build $(GOFLAGS) -o $(BINARY) cmd/mcp-server/main.go

test:
	$(GO) test $(GOFLAGS) ./...

test-phase-%:
	$(GO) test $(GOFLAGS) ./test/integration/$*_test.go

test-coverage:
	$(GO) test $(GOFLAGS) -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

run: build
	./$(BINARY)

clean:
	rm -f $(BINARY)
	rm -f coverage.out coverage.html
	rm -f test/deploy_runtime_tests

lint:
	$(GO) vet ./...
	$(GO) fmt ./...

fmt:
	$(GO) fmt ./...

# Development helpers
dev-test:
	ZEROPS_DEBUG=true $(GO) test $(GOFLAGS) -count=1 ./...

dev-run:
	ZEROPS_DEBUG=true $(GO) run cmd/mcp-server/main.go

# Phase-specific test targets
test-auth:
	$(GO) test $(GOFLAGS) -run TestAuth ./test/integration/

test-project:
	$(GO) test $(GOFLAGS) -run TestProject ./test/integration/

test-service:
	$(GO) test $(GOFLAGS) -run TestService ./test/integration/

test-deploy:
	$(GO) test $(GOFLAGS) -run TestDeploy ./test/integration/

test-config:
	$(GO) test $(GOFLAGS) -run TestConfig ./test/integration/

test-workflow:
	$(GO) test $(GOFLAGS) -run TestWorkflow ./test/integration/

# Cleanup test resources
cleanup-test:
	$(GO) run test/cleanup/main.go --prefix="mcp-test-"

# Runtime deployment tests
test-integration:
	$(GO) test $(GOFLAGS) ./test/integration -timeout $(TEST_TIMEOUT)

test-runtime: build
	@if [ -z "$(ZEROPS_API_KEY)" ]; then \
		echo "Error: ZEROPS_API_KEY not set"; \
		exit 1; \
	fi
	ZEROPS_TEST_FULL=true $(GO) test $(GOFLAGS) ./test/integration -run TestAllRuntimeDeployments -timeout $(TEST_TIMEOUT)

verify-runtime:
	$(GO) run test/deploy_runtime_tests.go

simulate-runtime:
	./scripts/deploy_runtime_tests.sh

test-report:
	./scripts/run_all_runtime_tests.sh

# Help
help:
	@echo "Zerops MCP Server v3 - Makefile targets:"
	@echo ""
	@echo "  make build          - Build the MCP server"
	@echo "  make test           - Run all tests"
	@echo "  make test-integration - Run integration tests"
	@echo "  make test-runtime   - Run full runtime deployment tests (requires ZEROPS_API_KEY)"
	@echo "  make verify-runtime - Verify runtime recipes are configured correctly"
	@echo "  make simulate-runtime - Simulate runtime deployments"
	@echo "  make run            - Build and run the MCP server"
	@echo "  make clean          - Remove build artifacts"
	@echo "  make dev-test       - Run development test script"
	@echo "  make fmt            - Format Go code"
	@echo "  make lint           - Run linter"
	@echo "  make help           - Show this help message"
	@echo ""
	@echo "Environment variables:"
	@echo "  ZEROPS_API_KEY      - Required for runtime tests"
	@echo "  ZEROPS_TEST_FULL    - Set to 'true' to run full deployment tests"
	@echo "  ZEROPS_TEST_PARALLEL - Set to 'true' to run tests in parallel"