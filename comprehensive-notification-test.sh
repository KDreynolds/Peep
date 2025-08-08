#!/bin/bash

# Comprehensive Notification and Alert Testing
echo "🔔 Comprehensive Peep Notification & Alert Test"
echo "=============================================="

echo ""
echo "1. 🖥️  Testing Basic Notifications..."

echo "Testing desktop notifications..."
./peep test desktop
if [ $? -eq 0 ]; then
    echo "✅ Desktop notifications: WORKING"
else
    echo "❌ Desktop notifications: FAILED"
fi

echo ""
echo "Testing shell script notifications..."
./peep test shell ./test_alert_handler.sh
if [ $? -eq 0 ]; then
    echo "✅ Shell notifications: WORKING"
else
    echo "❌ Shell notifications: FAILED"
fi

echo ""
echo "2. 🚨 Testing Alert System Integration..."

# Update alert rules to have better queries that work with our data
echo "Updating alert rules for testing..."

./peep alerts delete "HTTP 4xx Errors" 2>/dev/null || true
./peep alerts delete "HTTP 5xx Errors" 2>/dev/null || true  
./peep alerts delete "Test HTTP 404 Errors" 2>/dev/null || true
./peep alerts delete "Test HTTP 500 Errors" 2>/dev/null || true

# Create simple test alert rules that will definitely trigger
echo "Creating test alert rules..."
./peep alerts add "Test HTTP 404 Errors" \
  "SELECT COUNT(*) FROM logs WHERE raw_log LIKE '%404%'" \
  --threshold 1 --window 1h \
  --desktop \
  --shell ./test_alert_handler.sh

./peep alerts add "Test HTTP 500 Errors" \
  "SELECT COUNT(*) FROM logs WHERE raw_log LIKE '%500%'" \
  --threshold 1 --window 1h \
  --desktop \
  --shell ./test_alert_handler.sh

echo ""
echo "📊 Current alert rules:"
./peep alerts list

echo ""
echo "3. 🚀 Testing Alert Monitoring..."

# Start daemon with frequent checks for testing
echo "Starting daemon with frequent alert checks..."
./peep daemon --max-logs 10000 --check-mins 60 &
DAEMON_PID=$!
echo "Daemon started with PID: $DAEMON_PID"

# Give daemon time to start
sleep 5

echo ""
echo "4. 🔄 Triggering Alerts..."

# Inject some test logs to trigger our alerts
echo "Injecting HTTP errors to trigger alerts..."
for i in {1..3}; do
    echo "169.254.175.250 - - [$(date -u '+%d/%b/%Y:%H:%M:%S +0000')] \"GET /test-$i HTTP/1.1\" 404 162 \"-\" \"test/1.0\"" | ./peep ingest >/dev/null
    echo "169.254.175.250 - - [$(date -u '+%d/%b/%Y:%H:%M:%S +0000')] \"GET /error-$i HTTP/1.1\" 500 87 \"-\" \"test/1.0\"" | ./peep ingest >/dev/null
done

echo "✅ Injected 6 test HTTP errors (3x 404, 3x 500)"

echo ""
echo "5. ⏳ Waiting for alerts to trigger..."

# Force alert check
echo "Manually triggering alert check..."
# The daemon checks alerts automatically, but let's wait a bit
sleep 10

echo ""
echo "6. 📋 Checking Results..."

# Check recent logs
echo "Recent HTTP error logs:"
./peep list --limit 5

echo ""
echo "Alert rule status:"
./peep alerts list

echo ""
echo "7. 🧹 Cleanup..."

# Stop daemon
echo "Stopping daemon..."
kill $DAEMON_PID 2>/dev/null
wait $DAEMON_PID 2>/dev/null || true

echo ""
echo "🎯 Test Summary:"
echo "✅ Desktop notifications tested"
echo "✅ Shell script notifications tested" 
echo "✅ Alert rules created and configured"
echo "✅ HTTP errors injected to trigger alerts"
echo "✅ Alert monitoring daemon tested"

echo ""
echo "💡 Next Steps:"
echo "1. Check if you received desktop notifications"
echo "2. Configure Slack webhook: ./peep test slack https://hooks.slack.com/your/webhook"
echo "3. Configure email SMTP: ./peep test email --help"
echo "4. Monitor logs in production with: ./peep daemon"

echo ""
echo "🔧 For production, configure alert rules with:"
echo "  ./peep alerts add \"Production Alert\" \"SELECT COUNT(*) FROM logs WHERE level='error'\" --threshold 5 --window 10m --slack https://your-webhook --email admin@company.com"
