#!/bin/bash

# Peep Auto-Retention Demo
# Demonstrates automatic log retention and daemon mode

echo "🧹 Peep Auto-Retention Demo"
echo "=========================="

echo "📊 Current state:"
./peep stats

echo ""
echo "🚀 Testing daemon mode with auto-retention..."
echo "Settings: max-logs=500, check every 30 seconds"

# Start daemon in background for demo
./peep daemon --max-logs 500 --check-mins 1 --max-age-days 1 &
DAEMON_PID=$!
echo "Started daemon with PID: $DAEMON_PID"

# Let daemon start up
sleep 3

echo ""
echo "🔄 Adding logs to trigger retention..."

# Add some logs to exceed the threshold
for i in {1..100}; do
    echo "2025-08-07T$(date +%H:%M:%S)Z INFO demo-service Log message $i from retention demo" | ./peep ingest --exclude-levels debug >/dev/null 2>&1
done

echo "✅ Added 100 test logs"

echo ""
echo "📊 Stats after adding logs:"
./peep stats

echo ""
echo "⏳ Waiting for auto-retention to trigger (up to 60 seconds)..."

# Wait for retention to kick in
sleep 65

echo ""
echo "📊 Stats after auto-retention:"
./peep stats

echo ""
echo "🛑 Stopping daemon..."
kill $DAEMON_PID
wait $DAEMON_PID 2>/dev/null

echo ""
echo "🎯 Demo Summary:"
echo "- Auto-retention keeps database size manageable"
echo "- Configurable thresholds: log count, age, size"
echo "- Works in daemon mode and during ingestion"
echo "- Perfect for production deployment!"

echo ""
echo "💡 For production use:"
echo "  peep daemon --max-logs 100000 --max-age-days 30 --max-size-mb 1000"
