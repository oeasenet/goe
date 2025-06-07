#!/bin/bash

echo "=== Testing Fx Logs with INFO level ==="
echo "Creating .env with LOG_LEVEL=info"
cat > .env << 'EOF'
APP_NAME=FxTest
LOG_LEVEL=info
EOF

echo -e "\nRunning cache demo with INFO level..."
echo "Expected: Only INFO/WARN/ERROR logs, NO DEBUG logs"
echo "----------------------------------------"
go run cache_demo/main.go 2>&1 | head -20

echo -e "\n\n=== Testing Fx Logs with DEBUG level ==="
echo "Creating .env with LOG_LEVEL=debug"
cat > .env << 'EOF'
APP_NAME=FxTest
LOG_LEVEL=debug
EOF

echo -e "\nRunning cache demo with DEBUG level..."
echo "Expected: All logs including DEBUG (Fx provided, invoking, etc.)"
echo "----------------------------------------"
go run cache_demo/main.go 2>&1 | head -30

# Cleanup
rm -f .env

echo -e "\n✅ Test completed!"