#!/bin/bash

# Test script for unified-go server
# Tests health endpoint and basic functionality

PORT=5555
BASE_URL="http://localhost:$PORT"

echo "Starting unified-go server on port $PORT..."
PORT=$PORT ./unified-go > /tmp/unified-go-test.log 2>&1 &
SERVER_PID=$!

echo "Server started with PID: $SERVER_PID"
sleep 2

# Test health endpoint
echo ""
echo "Testing health endpoint..."
HEALTH_RESPONSE=$(curl -s "$BASE_URL/health")
if [ $? -eq 0 ]; then
    echo "✓ Health check successful"
    echo "$HEALTH_RESPONSE" | jq . 2>/dev/null || echo "$HEALTH_RESPONSE"
else
    echo "✗ Health check failed"
fi

# Test dashboard
echo ""
echo "Testing dashboard..."
DASHBOARD_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/dashboard")
if [ "$DASHBOARD_RESPONSE" = "200" ]; then
    echo "✓ Dashboard returns HTTP 200"
else
    echo "✗ Dashboard returned HTTP $DASHBOARD_RESPONSE"
fi

# Test typing app
echo ""
echo "Testing typing app..."
TYPING_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/typing")
if [ "$TYPING_RESPONSE" = "200" ]; then
    echo "✓ Typing app returns HTTP 200"
else
    echo "✗ Typing app returned HTTP $TYPING_RESPONSE"
fi

# Test API endpoints
echo ""
echo "Testing API endpoints..."
API_RESPONSE=$(curl -s "$BASE_URL/api/dashboard/stats")
if [ $? -eq 0 ]; then
    echo "✓ API endpoint working"
    echo "$API_RESPONSE" | jq . 2>/dev/null || echo "$API_RESPONSE"
else
    echo "✗ API endpoint failed"
fi

# Cleanup
echo ""
echo "Stopping server (PID: $SERVER_PID)..."
kill $SERVER_PID 2>/dev/null
sleep 1

echo ""
echo "Server logs:"
echo "----------------------------------------"
tail -20 /tmp/unified-go-test.log
echo "----------------------------------------"

echo ""
echo "Test complete!"
