#!/bin/bash

# Simple test script to verify the user service functionality
# This script tests the service without requiring Docker containers

echo "Testing User Service..."

# Start the service in the background
echo "Starting User Service..."
go run . &
SERVICE_PID=$!

# Wait for service to start
sleep 3

# Test health endpoint
echo "Testing health endpoint..."
curl -s http://localhost:8002/health | jq .

# Test user registration (this will fail without database, but we can check the endpoint exists)
echo "Testing user registration endpoint..."
curl -s -X POST http://localhost:8002/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","username":"testuser","first_name":"Test","last_name":"User","password":"Password123!"}' \
  | head -c 200

echo ""

# Clean up
echo "Stopping service..."
kill $SERVICE_PID

echo "Test completed!"