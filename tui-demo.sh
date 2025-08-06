#!/bin/bash
# TUI Demo Script - Shows real-time log ingestion with TUI

set -e

echo "🎪 Peep TUI Demo - Real-time Log Monitoring"
echo "==========================================="
echo

# Build the latest version
echo "🔨 Building Peep..."
make build > /dev/null 2>&1
echo

# Clean start
echo "🧹 Starting fresh..."
rm -f logs.db
echo

# Add some initial logs
echo "📥 Adding initial logs..."
echo '{"timestamp":"'$(date -u +%Y-%m-%dT%H:%M:%SZ)'","level":"info","message":"Application startup","service":"web-server"}' | ./peep > /dev/null
echo '{"timestamp":"'$(date -u +%Y-%m-%dT%H:%M:%SZ)'","level":"info","message":"Database connected","service":"db"}' | ./peep > /dev/null
echo '{"timestamp":"'$(date -u +%Y-%m-%dT%H:%M:%SZ)'","level":"warn","message":"High memory usage","service":"monitor"}' | ./peep > /dev/null
echo

echo "🖥️  Starting TUI in 3 seconds..."
echo "   Controls:"
echo "   • ↑/↓ or j/k  - Navigate logs"
echo "   • /           - Search mode"
echo "   • r           - Manual refresh"
echo "   • q           - Quit"
echo "   • esc         - Cancel search"
echo

# Background process to add logs while TUI is running
(
    sleep 5
    echo '{"timestamp":"'$(date -u +%Y-%m-%dT%H:%M:%SZ)'","level":"error","message":"Connection timeout","service":"api"}' | ./peep > /dev/null
    sleep 3
    echo '{"timestamp":"'$(date -u +%Y-%m-%dT%H:%M:%SZ)'","level":"info","message":"User authenticated","service":"auth"}' | ./peep > /dev/null
    sleep 3
    echo '{"timestamp":"'$(date -u +%Y-%m-%dT%H:%M:%SZ)'","level":"debug","message":"Cache hit","service":"cache"}' | ./peep > /dev/null
    sleep 3
    echo '{"timestamp":"'$(date -u +%Y-%m-%dT%H:%M:%SZ)'","level":"error","message":"Failed to save user data","service":"db"}' | ./peep > /dev/null
) &

sleep 3

# Start TUI
./peep tui

echo
echo "✅ TUI Demo complete!"
echo "💡 The TUI shows logs in real-time with:"
echo "   • Color-coded log levels"
echo "   • Auto-refresh every 2 seconds"
echo "   • Live search and filtering"
echo "   • Keyboard navigation"
