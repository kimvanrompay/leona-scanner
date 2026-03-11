#!/bin/bash
set -e

echo "🚀 LEONA Server Setup Script"
echo "=============================="

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Server IPs
PROD_IP="89.167.59.30"
DEV_IP="135.181.33.231"

# Function to setup a server
setup_server() {
    local SERVER_IP=$1
    local SERVER_TYPE=$2
    
    echo -e "${YELLOW}Setting up $SERVER_TYPE server ($SERVER_IP)...${NC}"
    
    ssh -o StrictHostKeyChecking=no root@$SERVER_IP << 'ENDSSH'
set -e

echo "📦 Updating system packages..."
apt-get update
apt-get upgrade -y

echo "🐳 Installing Docker..."
apt-get install -y ca-certificates curl
install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
chmod a+r /etc/apt/keyrings/docker.asc

echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
  tee /etc/apt/sources.list.d/docker.list > /dev/null

apt-get update
apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

echo "🔥 Setting up firewall..."
apt-get install -y ufw
ufw --force enable
ufw allow 22/tcp
ufw allow 80/tcp
ufw allow 443/tcp
ufw allow 8080:9000/tcp

echo "📝 Installing useful tools..."
apt-get install -y git htop ncdu vim tmux curl wget

echo "✅ Server setup complete!"
docker --version
docker compose version

ENDSSH
    
    echo -e "${GREEN}✅ $SERVER_TYPE server setup complete!${NC}"
}

# Setup production server
echo ""
echo -e "${YELLOW}=== PRODUCTION SERVER ===${NC}"
setup_server $PROD_IP "PRODUCTION"

# Setup development server
echo ""
echo -e "${YELLOW}=== DEVELOPMENT SERVER ===${NC}"
setup_server $DEV_IP "DEVELOPMENT"

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}🎉 Both servers are ready!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "Production: ssh root@$PROD_IP"
echo "Development: ssh root@$DEV_IP"
echo ""
echo "Next steps:"
echo "1. Deploy LEONA to production: ./deploy-prod.sh"
echo "2. Access dev server: ssh root@$DEV_IP"
