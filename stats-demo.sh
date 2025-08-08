#!/bin/bash

# Peep Stats Monitoring Demo
# Example script showing how to use peep stats for system monitoring

echo "🔍 Peep Health Check Demo"
echo "========================="

# Basic health check
echo "📊 Current Stats:"
./peep stats

echo ""
echo "📈 JSON Stats for Scripting:"
STATS=$(./peep stats --json)

# Extract specific metrics using jq (if available)
if command -v jq &> /dev/null; then
    echo "Total logs: $(echo $STATS | jq '.total_logs')"
    echo "Database size: $(echo $STATS | jq '.database_size_mb') MB"
    echo "Memory usage: $(echo $STATS | jq '.memory_usage_mb') MB"
    echo "Active alerts: $(echo $STATS | jq '.active_alert_rules')"
    
    # Check if database is getting too large
    DB_SIZE=$(echo $STATS | jq '.database_size_mb')
    if (( $(echo "$DB_SIZE > 100" | bc -l) )); then
        echo "⚠️  WARNING: Database size ($DB_SIZE MB) exceeds recommended limit!"
        echo "💡 Consider running: ./peep clean --keep-last 10000"
    fi
    
    # Check memory usage
    MEM_USAGE=$(echo $STATS | jq '.memory_usage_mb')
    if (( $(echo "$MEM_USAGE > 50" | bc -l) )); then
        echo "⚠️  WARNING: High memory usage ($MEM_USAGE MB)"
    fi
else
    echo "💡 Install 'jq' for advanced JSON parsing"
fi

echo ""
echo "🚀 This could be run as a cron job for monitoring!"
echo "Example crontab entry:"
echo "*/5 * * * * /path/to/peep-monitor.sh"
