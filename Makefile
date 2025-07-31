# Minecraft Mod Manager - Development Makefile
# Provides convenient commands for development, testing, and building

.PHONY: help install install-dev test lint type-check format clean build build-exe health-check run-tests coverage docs

# Default target
help:
	@echo "🎮 Minecraft Mod Manager - Development Commands"
	@echo "==============================================="
	@echo ""
	@echo "Setup Commands:"
	@echo "  install      - Install package in development mode"
	@echo "  setup-dev    - Full development environment setup
  install-dev  - Install with development dependencies"
	@echo ""
	@echo "Quality Assurance:"
	@echo "  test         - Run all tests"
	@echo "  lint         - Run code linting (ruff)"
	@echo "  type-check   - Run type checking (mypy)"
	@echo "  format       - Format code (ruff)"
	@echo "  coverage     - Run tests with coverage report"
	@echo ""
	@echo "Build Commands:"
	@echo "  build        - Build wheel package"
	@echo "  build-exe    - Build standalone executable"
	@echo "  clean        - Clean build artifacts"
	@echo ""
	@echo "Application Commands:"
	@echo "  health-check - Run system health checks"
	@echo "  cleanup      - Run system cleanup"
	@echo ""
	@echo "Development:"
	@echo "  run-tests    - Run tests with verbose output"
	@echo "  docs         - Generate documentation"

# Setup commands
install:
	@echo "📦 Installing package in development mode..."
	pip install -e .

install-dev:
	@echo "🛠️  Installing with development dependencies..."
	pip install -e .
	pip install -r requirements-dev.txt

# Quality assurance
test:
	@echo "🧪 Running tests..."
	python -m pytest tests/ -v

lint:
	@echo "🔍 Running code linting..."
	python -m ruff check minecraft_mod_manager/

type-check:
	@echo "🔍 Running type checking..."
	python -m mypy minecraft_mod_manager/

format:
	@echo "✨ Formatting code..."
	python -m ruff format minecraft_mod_manager/
	python -m ruff check --fix minecraft_mod_manager/

coverage:
	@echo "📊 Running tests with coverage..."
	python -m pytest tests/ --cov=minecraft_mod_manager --cov-report=html --cov-report=term

# Build commands
build:
	@echo "🏗️  Building wheel package..."
	./package-build.sh

build-exe:
	@echo "🏗️  Building standalone executable..."
	./executable-build.sh

clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf build/
	rm -rf dist/
	rm -rf *.egg-info/
	rm -rf .pytest_cache/
	rm -rf .mypy_cache/
	rm -rf .ruff_cache/
	rm -rf htmlcov/
	find . -type d -name __pycache__ -exec rm -rf {} +
	find . -type f -name "*.pyc" -delete
	find . -type f -name "*.pyo" -delete
	find . -type f -name "*.coverage" -delete

# Application commands
health-check:
	@echo "🏥 Running system health checks..."
	minecraft-mod-manager --health-check

cleanup:
	@echo "🧹 Running system cleanup..."
	minecraft-mod-manager --cleanup

# Development commands
run-tests:
	@echo "🧪 Running tests with verbose output..."
	python -m pytest tests/ -v -s

docs:
	@echo "📚 Generating documentation..."
	@echo "Documentation generation not yet implemented"

# Quality check all
check-all: lint type-check test
	@echo "✅ All quality checks passed!"

# Full development setup
setup-dev:
	@echo "🚀 Running full development setup..."
	./scripts/development-setup.sh