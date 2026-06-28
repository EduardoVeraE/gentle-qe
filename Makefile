# Gentle-QE — tareas de desarrollo del fork SDET.
# El overlay SDET vive sobre upstream gentle-ai; `verify-overlay` es el guard
# que CI corre como job `overlay-guard` (.github/workflows/ci.yml).

.DEFAULT_GOAL := help
.PHONY: help build test verify-overlay verify

help: ## Lista los targets disponibles
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

build: ## Compila todos los paquetes
	go build ./...

test: ## Corre la suite de tests
	go test ./...

verify-overlay: ## Verifica integridad del overlay SDET y drift vs upstream
	go run ./tools/qe-overlay verify

verify: build test verify-overlay ## build + test + verify-overlay (paridad con CI)
