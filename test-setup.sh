#!/bin/bash

echo "================================================"
echo "Shopware Analytics - Setup Verification"
echo "================================================"
echo ""

# Check Go
echo "Checking Go installation..."
if command -v go &> /dev/null; then
    echo "✓ Go $(go version | awk '{print $3}') installed"
else
    echo "✗ Go not installed"
    exit 1
fi

# Check PHP
echo "Checking PHP installation..."
if command -v php &> /dev/null; then
    echo "✓ PHP $(php -v | head -n1 | awk '{print $2}') installed"
else
    echo "✗ PHP not installed"
    exit 1
fi

# Check Node
echo "Checking Node.js installation..."
if command -v node &> /dev/null; then
    echo "✓ Node.js $(node --version) installed"
else
    echo "✗ Node.js not installed"
    exit 1
fi

# Check Docker
echo "Checking Docker installation..."
if command -v docker &> /dev/null; then
    echo "✓ Docker $(docker --version | awk '{print $3}' | tr -d ',') installed"
else
    echo "✗ Docker not installed"
    exit 1
fi

# Check Docker Compose
echo "Checking Docker Compose..."
if command -v docker-compose &> /dev/null || docker compose version &> /dev/null; then
    echo "✓ Docker Compose available"
else
    echo "✗ Docker Compose not available"
    exit 1
fi

echo ""
echo "================================================"
echo "All prerequisites installed!"
echo "================================================"
echo ""
echo "Project Structure:"
tree -L 2 -I 'node_modules|vendor' . || ls -R

echo ""
echo "To start the application:"
echo "  make build && make up"
echo ""
echo "Or:"
echo "  docker-compose build && docker-compose up -d"
