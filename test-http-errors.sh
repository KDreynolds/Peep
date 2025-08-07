#!/bin/bash

# HTTP Error Detection Test Script
# Simulates various HTTP error scenarios for testing alert rules

echo "🧪 Testing HTTP Error Detection..."

# Simulate HTTP 404 errors
echo "169.254.175.250 - - [$(date -u '+%d/%b/%Y:%H:%M:%S +0000')] \"GET /missing-page HTTP/1.1\" 404 162 \"-\" \"curl/7.68.0\"" | ./peep ingest

echo "169.254.175.250 - - [$(date -u '+%d/%b/%Y:%H:%M:%S +0000')] \"GET /api/v1/missing HTTP/1.1\" 404 162 \"-\" \"PostmanRuntime/7.29.0\"" | ./peep ingest

# Simulate HTTP 500 errors
echo "169.254.175.250 - - [$(date -u '+%d/%b/%Y:%H:%M:%S +0000')] \"POST /api/process HTTP/1.1\" 500 87 \"-\" \"axios/0.21.1\"" | ./peep ingest

echo "169.254.175.250 - - [$(date -u '+%d/%b/%Y:%H:%M:%S +0000')] \"GET /dashboard HTTP/1.1\" 500 87 \"-\" \"Mozilla/5.0\"" | ./peep ingest

# Simulate HTTP 403 errors
echo "169.254.175.250 - - [$(date -u '+%d/%b/%Y:%H:%M:%S +0000')] \"GET /admin HTTP/1.1\" 403 146 \"-\" \"curl/7.68.0\"" | ./peep ingest

# Simulate HTTP 503 errors (service unavailable)
echo "169.254.175.250 - - [$(date -u '+%d/%b/%Y:%H:%M:%S +0000')] \"GET /api/health HTTP/1.1\" 503 95 \"-\" \"kube-probe/1.32+\"" | ./peep ingest

echo "✅ Injected 6 HTTP error logs (2x 404, 2x 500, 1x 403, 1x 503)"

# Give a moment for the logs to be processed
sleep 1

# Check if our error detection queries work
echo ""
echo "🔍 Testing HTTP Error Detection Queries:"

echo "4xx Errors Found:"
sqlite3 logs.db "SELECT COUNT(*) FROM logs WHERE raw_log LIKE '%HTTP/1.1\" 4%' AND timestamp > datetime('now', '-5 minutes');"

echo "5xx Errors Found:"
sqlite3 logs.db "SELECT COUNT(*) FROM logs WHERE raw_log LIKE '%HTTP/1.1\" 5%' AND timestamp > datetime('now', '-5 minutes');"

echo "Total HTTP Errors Found:"
sqlite3 logs.db "SELECT COUNT(*) FROM logs WHERE (raw_log LIKE '%HTTP/1.1\" 4%' OR raw_log LIKE '%HTTP/1.1\" 5%') AND timestamp > datetime('now', '-5 minutes');"

echo ""
echo "📋 Recent HTTP Error Logs:"
sqlite3 logs.db "SELECT timestamp, raw_log FROM logs WHERE (raw_log LIKE '%HTTP/1.1\" 4%' OR raw_log LIKE '%HTTP/1.1\" 5%') AND timestamp > datetime('now', '-5 minutes') ORDER BY timestamp DESC LIMIT 5;"
