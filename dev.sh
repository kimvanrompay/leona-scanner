#!/bin/bash

# LEONA Development Server
# This script starts the development server with SQLite

echo "🚀 Starting LEONA CRA Scanner (Development Mode)"
echo "📊 Using SQLite database: ./leona.db"
echo ""

# Check if .env exists
if [ ! -f .env ]; then
    echo "⚠️  Warning: .env file not found!"
    echo "📝 Creating .env from .env.example..."
    cp .env.example .env
    echo "✅ Created .env file. Please edit it with your credentials."
    echo ""
fi

# Install dependencies if needed
if [ ! -d "vendor" ]; then
    echo "📦 Installing Go dependencies..."
    go mod download
fi

# Build and run
echo "🔨 Building application..."
go build -o leona-scanner cmd/server/main.go

if [ $? -eq 0 ]; then
    echo "✅ Build successful!"
    echo ""
    echo "Starting server..."
    ./leona-scanner
else
    echo "❌ Build failed!"
    exit 1
fi
