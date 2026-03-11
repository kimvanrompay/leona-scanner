#!/bin/bash
# Setup SSH keys for passwordless login to LEONA servers

echo "🔑 Setting up SSH keys for LEONA servers"
echo "=========================================="
echo ""

# Your public key
PUBLIC_KEY="ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIDvGt6yyZ5LybkbmkmlCDmMg4CMTC9zOFP5ricqVF9De kim.vanrompay@livecadia.com"

# Server details
PROD_IP="89.167.59.30"
PROD_PASS="fmsaaefJHur4dueVKVCf"
DEV_IP="135.181.33.231"
DEV_PASS="7JjpMPicNXgHWJxFwTRs"

setup_key() {
    local SERVER_IP=$1
    local SERVER_NAME=$2
    
    echo "📡 Setting up SSH key for $SERVER_NAME ($SERVER_IP)..."
    echo ""
    echo "When prompted, enter the password for this server."
    echo "After this, you'll never need to enter it again!"
    echo ""
    
    # Add key via ssh-copy-id (will prompt for password)
    ssh-copy-id -i ~/.ssh/id_ed25519.pub root@$SERVER_IP
    
    if [ $? -eq 0 ]; then
        echo "✅ SSH key added to $SERVER_NAME!"
        echo "Test connection (should work without password):"
        ssh root@$SERVER_IP 'echo "✅ Passwordless SSH works!"'
    else
        echo "❌ Failed to add key to $SERVER_NAME"
    fi
    echo ""
}

echo "Server passwords (copy these for when prompted):"
echo "Production ($PROD_IP): $PROD_PASS"
echo "Development ($DEV_IP): $DEV_PASS"
echo ""
read -p "Press Enter to continue..."

# Setup production server
setup_key $PROD_IP "Production"

# Setup development server
setup_key $DEV_IP "Development"

echo ""
echo "=========================================="
echo "🎉 Setup complete!"
echo "=========================================="
echo ""
echo "You can now connect without passwords:"
echo "  ssh root@$PROD_IP   (Production)"
echo "  ssh root@$DEV_IP    (Development)"
echo ""
echo "Or use the Warp workflows you just created!"
