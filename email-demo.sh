#!/bin/bash

echo "📧 Peep Email Notifications Demo"
echo "================================="
echo ""
echo "This script demonstrates how to configure and test email notifications in Peep."
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}1. Adding an email notification channel:${NC}"
echo ""
echo "peep alerts channels add email \"Team Alerts\" \\"
echo "  --smtp-host smtp.gmail.com \\"
echo "  --smtp-port 587 \\"
echo "  --username your-email@gmail.com \\"
echo "  --password your-app-password \\"
echo "  --from your-email@gmail.com \\"
echo "  --from-name \"Peep Alerts\" \\"
echo "  --to team@company.com,admin@company.com"
echo ""

echo -e "${BLUE}2. Testing email configuration:${NC}"
echo ""
echo "peep test email \\"
echo "  --smtp-host smtp.gmail.com \\"
echo "  --username your-email@gmail.com \\"
echo "  --password your-app-password \\"
echo "  --from your-email@gmail.com \\"
echo "  --to recipient@example.com"
echo ""

echo -e "${BLUE}3. Email notification features:${NC}"
echo "  ✅ Rich HTML email formatting"
echo "  ✅ Severity-based color coding (Critical=Red, Warning=Orange, Info=Blue)"
echo "  ✅ Professional email templates"
echo "  ✅ Multiple recipients supported"
echo "  ✅ Customizable from name and email"
echo "  ✅ Support for Gmail, Outlook, and other SMTP providers"
echo ""

echo -e "${BLUE}4. Common SMTP configurations:${NC}"
echo ""
echo -e "${YELLOW}Gmail:${NC}"
echo "  --smtp-host smtp.gmail.com --smtp-port 587"
echo "  Note: Use app password, not your regular password"
echo ""
echo -e "${YELLOW}Outlook/Hotmail:${NC}"
echo "  --smtp-host smtp-mail.outlook.com --smtp-port 587"
echo ""
echo -e "${YELLOW}Yahoo:${NC}"
echo "  --smtp-host smtp.mail.yahoo.com --smtp-port 587"
echo ""

echo -e "${BLUE}5. Example alert rule that triggers email:${NC}"
echo ""
echo "peep alerts add \"Critical Errors\" \\"
echo "  \"SELECT COUNT(*) FROM logs WHERE level='error' AND timestamp > datetime('now', '-5 minutes')\" \\"
echo "  --threshold 5 --window 5m"
echo ""

echo -e "${GREEN}📧 Email notifications are ready to keep your team informed!${NC}"
echo "   • Professional HTML formatting"
echo "   • Severity-based visual indicators"  
echo "   • Reliable SMTP delivery"
echo "   • Team collaboration ready"
