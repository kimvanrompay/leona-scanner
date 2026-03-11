#!/bin/bash
# Quick test script for email delivery

echo "🧪 Testing LEONA email system..."
echo ""
echo "Step 1: Checking .env configuration..."

if ! grep -q "SMTP_HOST=mail1.netim.hosting" .env; then
    echo "❌ Wrong SMTP_HOST in .env"
    echo "   Expected: mail1.netim.hosting"
    echo "   Run: Update your .env file"
    exit 1
fi

if ! grep -q "SMTP_PORT=465" .env; then
    echo "❌ Wrong SMTP_PORT in .env"
    echo "   Expected: 465"
    exit 1
fi

if grep -q "SMTP_PASS=your-app-specific-password" .env; then
    echo "⚠️  SMTP_PASS not configured!"
    echo ""
    echo "Please update .env with your actual email password:"
    echo "  SMTP_PASS=your_actual_password"
    echo ""
    echo "Your password is the same one you use for:"
    echo "  https://mail1.netim.hosting/webmail/"
    echo ""
    exit 1
fi

echo "✅ .env configuration looks good"
echo ""
echo "Step 2: Starting server..."
echo ""

# Start server in background
go run cmd/server/main.go &
SERVER_PID=$!

# Wait for server to start
sleep 3

echo ""
echo "Step 3: Testing engineer lead magnet..."
echo ""
echo "Visit: http://localhost:8080"
echo "Scroll to: 'Voor Engineers en Juristen'"
echo "Enter your email and click 'Ontvang Layer'"
echo ""
echo "Expected result:"
echo "  ✅ Green success message"
echo "  📧 Email in your inbox within 30 seconds"
echo "  Subject: 'Jouw meta-leona CRA Validator Layer'"
echo ""
echo "Press Ctrl+C when done testing"
echo ""

# Keep script running
wait $SERVER_PID
