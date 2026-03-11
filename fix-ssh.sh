#!/bin/bash
# Fix SSH access by adding key directly to both servers

echo "🔧 Adding SSH keys to LEONA servers via Hetzner..."

# Read your public key
PUB_KEY=$(cat ~/.ssh/id_ed25519.pub)

# For production server
echo "Adding key to production server..."
hcloud server ssh leona-prod "mkdir -p ~/.ssh && chmod 700 ~/.ssh && echo '$PUB_KEY' >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys && echo '✅ Production SSH key added!'"

# For development server  
echo "Adding key to development server..."
hcloud server ssh leona-dev "mkdir -p ~/.ssh && chmod 700 ~/.ssh && echo '$PUB_KEY' >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys && echo '✅ Development SSH key added!'"

echo ""
echo "🎉 Done! Testing connections..."
echo ""

# Test production
echo "Testing production..."
ssh -o ConnectTimeout=5 root@89.167.59.30 'echo "✅ Production works!"' && echo "Success!" || echo "❌ Failed"

# Test development
echo "Testing development..."
ssh -o ConnectTimeout=5 root@135.181.33.231 'echo "✅ Development works!"' && echo "Success!" || echo "❌ Failed"
