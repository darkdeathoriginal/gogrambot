#!/bin/sh

# Loop forever
while true; do
    echo "🔨 Building binary..."
    if [ -f "prebuild.go" ]; then
        echo "⚙️ Running prebuild.go..."
        go run prebuild.go
    fi
    # Build the bot, outputting to a binary named 'bot'
    go mod tidy
    CGO_ENABLED=1 go build -o bot main.go
    
    # Check if build succeeded
    if [ $? -eq 0 ]; then
        echo "✅ Build successful. Starting Bot..."
        # Run the bot
        ./bot
        
        # If bot crashes or stops (via /restart), we loop back and rebuild
        echo "⚠️ Bot stopped. Recompiling in 2 seconds..."
        sleep 2
    else
        echo "❌ Build failed. Waiting 10 seconds before retrying..."
        sleep 10
    fi
done