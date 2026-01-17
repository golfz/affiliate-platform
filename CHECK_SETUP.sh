#!/bin/bash

# Setup Check Script for Jenosize Affiliate Platform
# This script checks if all prerequisites are installed and ready

echo "🔍 Checking Setup for Jenosize Affiliate Platform..."
echo "=================================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

ERRORS=0
WARNINGS=0

# Function to check command
check_command() {
    local cmd=$1
    local name=$2
    local required=$3
    
    if command -v $cmd &> /dev/null; then
        version=$($cmd --version 2>/dev/null | head -1)
        echo -e "${GREEN}✓${NC} $name: INSTALLED ($version)"
        return 0
    else
        if [ "$required" = "required" ]; then
            echo -e "${RED}✗${NC} $name: NOT INSTALLED (REQUIRED)"
            ERRORS=$((ERRORS + 1))
            return 1
        else
            echo -e "${YELLOW}⚠${NC} $name: NOT INSTALLED (OPTIONAL)"
            WARNINGS=$((WARNINGS + 1))
            return 1
        fi
    fi
}

# Check Go
echo "📦 Prerequisites:"
check_command "go" "Go" "required"

# Check Node.js
check_command "node" "Node.js" "required"

# Check npm
check_command "npm" "npm" "required"

# Check Docker
check_command "docker" "Docker" "required"

# Check Docker Compose
check_command "docker-compose" "Docker Compose" "required"

# Check Make
check_command "make" "Make" "required"

echo ""
echo "📁 Project Files:"
echo "-----------------"

# Check go.mod
if [ -f "go.mod" ]; then
    echo -e "${GREEN}✓${NC} go.mod: EXISTS"
else
    echo -e "${RED}✗${NC} go.mod: NOT FOUND"
    ERRORS=$((ERRORS + 1))
fi

# Check package.json
if [ -f "apps/web/package.json" ]; then
    echo -e "${GREEN}✓${NC} package.json: EXISTS"
else
    echo -e "${YELLOW}⚠${NC} package.json: NOT FOUND (Frontend may not be set up)"
    WARNINGS=$((WARNINGS + 1))
fi

# Check config.json
if [ -f "configs/config.json" ]; then
    echo -e "${GREEN}✓${NC} config.json: EXISTS"
else
    if [ -f "configs/config.example.json" ]; then
        echo -e "${YELLOW}⚠${NC} config.json: NOT FOUND (will be created on 'make init')"
        WARNINGS=$((WARNINGS + 1))
    else
        echo -e "${RED}✗${NC} config.json: NOT FOUND (config.example.json also missing)"
        ERRORS=$((ERRORS + 1))
    fi
fi

# Check Swagger docs
if [ -f "docs/docs.go" ]; then
    echo -e "${GREEN}✓${NC} Swagger docs: GENERATED"
else
    echo -e "${YELLOW}⚠${NC} Swagger docs: NOT GENERATED (run 'make swagger')"
    WARNINGS=$((WARNINGS + 1))
fi

echo ""
echo "🐳 Docker Services:"
echo "-------------------"

# Check Docker services
if docker-compose ps &> /dev/null; then
    RUNNING=$(docker-compose ps --services --filter "status=running" 2>/dev/null | wc -l | xargs)
    TOTAL=$(docker-compose ps --services 2>/dev/null | wc -l | xargs)
    
    if [ "$TOTAL" -gt 0 ]; then
        if [ "$RUNNING" -eq "$TOTAL" ] && [ "$TOTAL" -gt 0 ]; then
            echo -e "${GREEN}✓${NC} Docker services: RUNNING ($RUNNING/$TOTAL)"
        else
            echo -e "${YELLOW}⚠${NC} Docker services: PARTIALLY RUNNING ($RUNNING/$TOTAL)"
            echo "   Run 'docker-compose up -d' to start services"
            WARNINGS=$((WARNINGS + 1))
        fi
    else
        echo -e "${YELLOW}⚠${NC} Docker services: NOT DEFINED or NOT RUNNING"
        echo "   Run 'make init' or 'docker-compose up -d' to start services"
        WARNINGS=$((WARNINGS + 1))
    fi
else
    echo -e "${RED}✗${NC} Docker Compose: UNABLE TO CHECK"
    ERRORS=$((ERRORS + 1))
fi

echo ""
echo "📦 Dependencies:"
echo "----------------"

# Check Go modules
if [ -f "go.mod" ]; then
    if [ -d "$(go env GOPATH)/pkg/mod" ] || [ -f "go.sum" ]; then
        echo -e "${GREEN}✓${NC} Go modules: AVAILABLE (check 'go mod tidy' if issues)"
    else
        echo -e "${YELLOW}⚠${NC} Go modules: MAY NEED 'go mod download'"
        WARNINGS=$((WARNINGS + 1))
    fi
else
    echo -e "${RED}✗${NC} Go modules: CANNOT CHECK (go.mod not found)"
fi

# Check Node modules
if [ -d "apps/web/node_modules" ]; then
    echo -e "${GREEN}✓${NC} Node modules: INSTALLED"
else
    echo -e "${YELLOW}⚠${NC} Node modules: NOT INSTALLED (will be installed on 'make init')"
    WARNINGS=$((WARNINGS + 1))
fi

echo ""
echo "🔧 CI/CD:"
echo "--------"

# Check CI workflows
if [ -d ".github/workflows" ]; then
    WORKFLOWS=$(ls -1 .github/workflows/*.yml .github/workflows/*.yaml 2>/dev/null | wc -l | xargs)
    if [ "$WORKFLOWS" -gt 0 ]; then
        echo -e "${GREEN}✓${NC} CI workflows: CONFIGURED ($WORKFLOWS files)"
    else
        echo -e "${YELLOW}⚠${NC} CI workflows: NOT CONFIGURED"
        WARNINGS=$((WARNINGS + 1))
    fi
else
    echo -e "${YELLOW}⚠${NC} CI workflows: DIRECTORY NOT FOUND"
    WARNINGS=$((WARNINGS + 1))
fi

# Check linter config
if [ -f ".golangci.yml" ]; then
    echo -e "${GREEN}✓${NC} Linter config: CONFIGURED"
else
    echo -e "${YELLOW}⚠${NC} Linter config: NOT FOUND"
    WARNINGS=$((WARNINGS + 1))
fi

echo ""
echo "=================================================="

# Summary
if [ $ERRORS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    echo -e "${GREEN}✅ All checks passed! Ready to run.${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. make init    # Initialize project (first time)"
    echo "  2. make mu      # Run database migrations"
    echo "  3. make seed    # Seed demo data (optional)"
    echo "  4. make swagger # Generate Swagger docs"
    echo "  5. make start   # Start backend + frontend"
    exit 0
elif [ $ERRORS -eq 0 ]; then
    echo -e "${YELLOW}⚠ $WARNINGS warning(s) found (non-critical)${NC}"
    echo ""
    echo "You can proceed, but consider fixing warnings."
    echo ""
    echo "Quick fixes:"
    echo "  - Run 'make init' to set up config and install dependencies"
    echo "  - Run 'make swagger' to generate Swagger docs"
    echo "  - Run 'docker-compose up -d' if services not running"
    exit 0
else
    echo -e "${RED}❌ $ERRORS error(s) found (must fix before running)${NC}"
    echo ""
    echo "Please fix the errors above before proceeding."
    exit 1
fi
