#!/bin/zsh

# Build the TCP proxy
echo "Building TCP proxy..."
go build -o tcp-proxy .

# Check if build was successful
if [ $? -eq 0 ]; then
    echo "Build successful! Executable created: tcp-proxy"
else
    echo "Build failed"
    exit 1
fi
