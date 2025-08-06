#!/bin/bash
# Slack Integration Demo for Peep

set -e

echo "📱 Peep Slack Integration Demo"
echo "=============================="
echo

# Build the latest version
echo "🔨 Building Peep..."
make build > /dev/null 2>&1
echo

echo "🎯 This demo shows Peep's Slack integration capabilities:"
echo "   • Rich Slack notifications with colors and attachments"
echo "   • Webhook configuration and management"
echo "   • Alert severity levels (Low/Medium/High/Critical)"
echo "   • Secure webhook URL masking"
echo

echo "📋 Current notification channels:"
./peep alerts channels list
echo

echo "🚨 Current alert rules:"
./peep alerts list
echo

echo "💡 To test with a real Slack channel:"
echo "   1. Create a Slack webhook at https://api.slack.com/incoming-webhooks"
echo "   2. Run: peep alerts channels add slack \"Your Team\" --webhook YOUR_WEBHOOK_URL"
echo "   3. Run: peep test slack YOUR_WEBHOOK_URL"
echo "   4. Start monitoring: peep alerts start"
echo

echo "🎉 Slack Integration Features:"
echo "   ✅ Rich message formatting with colors"
echo "   ✅ Alert severity indicators (🔴🟠🟡🟢)"
echo "   ✅ Structured data (count, threshold, timestamp)"
echo "   ✅ Secure webhook URL storage"
echo "   ✅ Multiple channel support"
echo "   ✅ Test commands for validation"
echo

echo "📱 Example Slack message format:"
echo "   🚨 Alert: High Error Rate"
echo "   Alert threshold exceeded: 5 events detected (limit: 3)"
echo "   Count: 5 | Threshold: 3 | Severity: 🟠 High"
echo "   Footer: Peep Observability | Timestamp: $(date)"
echo

echo "✨ Ready for team observability!"
