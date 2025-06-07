#!/bin/bash

# Start the server in background
echo "Starting cache HTTP demo server..."
go run main.go &
SERVER_PID=$!

# Wait for server to start
sleep 3

echo -e "\n=== Testing Cache HTTP Demo ==="

# Test home page
echo -e "\n1. Testing home page:"
curl -s http://localhost:8080/ | grep -q "Cache HTTP Demo" && echo "✅ Home page works" || echo "❌ Home page failed"

# Test get product (cache miss)
echo -e "\n2. Testing GET /api/products/1 (cache miss):"
RESPONSE=$(curl -s http://localhost:8080/api/products/1)
echo "$RESPONSE" | jq '.' 2>/dev/null || echo "$RESPONSE"

# Test get product again (cache hit)
echo -e "\n3. Testing GET /api/products/1 (cache hit):"
RESPONSE=$(curl -s http://localhost:8080/api/products/1)
echo "$RESPONSE" | jq '.' 2>/dev/null || echo "$RESPONSE"

# Test create product
echo -e "\n4. Testing POST /api/products:"
RESPONSE=$(curl -s -X POST http://localhost:8080/api/products \
  -H "Content-Type: application/json" \
  -d '{"id":2,"name":"Gaming Mouse","price":79.99,"stock":50}')
echo "$RESPONSE" | jq '.' 2>/dev/null || echo "$RESPONSE"

# Test cache stats
echo -e "\n5. Testing GET /api/stats:"
RESPONSE=$(curl -s http://localhost:8080/api/stats)
echo "$RESPONSE" | jq '.' 2>/dev/null || echo "$RESPONSE"

# Test clear cache
echo -e "\n6. Testing DELETE /api/cache:"
RESPONSE=$(curl -s -X DELETE http://localhost:8080/api/cache)
echo "$RESPONSE" | jq '.' 2>/dev/null || echo "$RESPONSE"

# Test cache stats after clear
echo -e "\n7. Testing GET /api/stats (after clear):"
RESPONSE=$(curl -s http://localhost:8080/api/stats)
echo "$RESPONSE" | jq '.' 2>/dev/null || echo "$RESPONSE"

# Kill the server
echo -e "\n=== Stopping server ==="
kill $SERVER_PID
wait $SERVER_PID 2>/dev/null

echo -e "\n✅ All tests completed!"