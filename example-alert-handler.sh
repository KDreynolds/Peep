#!/bin/bash

# Peep Alert Handler Example Script
# This script receives alert information via environment variables

echo "🚨 Peep Alert Received!"
echo "======================="
echo "Title: $PEEP_ALERT_TITLE"
echo "Severity: $PEEP_ALERT_SEVERITY"
echo "Count: $PEEP_ALERT_COUNT"
echo "Threshold: $PEEP_ALERT_THRESHOLD"
echo "Ratio: $PEEP_ALERT_RATIO"
echo "Timestamp: $PEEP_ALERT_TIMESTAMP"
echo ""
echo "Message:"
echo "$PEEP_ALERT_MESSAGE"
echo ""

# Example: Log to a file
echo "$(date): Alert - $PEEP_ALERT_TITLE ($PEEP_ALERT_COUNT/$PEEP_ALERT_THRESHOLD)" >> /tmp/peep-alerts.log

# Example: Send to a webhook (uncomment to use)
# curl -X POST https://your-webhook-url.com/alerts \
#   -H "Content-Type: application/json" \
#   -d "{\"title\":\"$PEEP_ALERT_TITLE\",\"severity\":\"$PEEP_ALERT_SEVERITY\",\"count\":$PEEP_ALERT_COUNT}"

# Example: Play a sound (macOS)
# if command -v afplay &> /dev/null; then
#   afplay /System/Library/Sounds/Glass.aiff
# fi

# Example: Send system notification (Linux)
# if command -v notify-send &> /dev/null; then
#   notify-send "Peep Alert" "$PEEP_ALERT_TITLE: $PEEP_ALERT_COUNT events"
# fi

echo "✅ Alert handled successfully!"
