# LoyalCoin Installation & Deployment Guide

Complete guide for installing, configuring, and deploying the LoyalCoin loyalty rewards platform.

---

## 📋 **Table of Contents**

1. [Prerequisites](#prerequisites)
2. [Local Development Setup](#local-development-setup)
3. [Infrastructure Setup](#infrastructure-setup)
4. [Backend Deployment](#backend-deployment)
5. [Frontend Deployment](#frontend-deployment)
6. [Production Deployment](#production-deployment)
7. [Monitoring & Maintenance](#monitoring--maintenance)
8. [Troubleshooting](#troubleshooting)

---

## 🔧 **Prerequisites**

### **Required Software**

| Software | Version | Purpose |
|----------|---------|---------|
| **Node.js** | 18.x or higher | Frontend build & runtime |
| **npm** | 9.x or higher | Package management |
| **Go** | 1.24 or higher | Backend server |
| **Docker** | 24.x or higher | Containerization |
| **Docker Compose** | 2.x or higher | Multi-container orchestration |
| **MongoDB** | 7.0+ | Database |
| **Git** | 2.x+ | Version control |

### **External Services**

1. **Blockfrost Account**: [blockfrost.io](https://blockfrost.io)
   - Sign up for free
   - Create a project for Cardano Preprod Testnet
   - Save your Project ID

2. **Cardano Wallet** (for admin):
   - Install [Eternl](https://eternl.io) or [Nami](https://namiwallet.io)
   - Switch to Preprod Testnet
   - Fund wallet with test ADA from [faucet](https://docs.cardano.org/cardano-testnets/tools/faucet/)

### **System Requirements**

**Minimum:**
- CPU: 2 cores
- RAM: 4 GB
- Storage: 20 GB
- OS: Linux (Ubuntu 22.04 recommended), macOS, or Windows with WSL2

**Recommended for Production:**
- CPU: 4+ cores
- RAM: 8+ GB
- Storage: 100+ GB SSD
- OS: Ubuntu Server 22.04 LTS

---

## 🏠 **Local Development Setup**

### **Step 1: Clone Repository**

```bash
git clone https://github.com/yourorg/loyalcoin.git
cd loyalcoin
```

### **Step 2: Start Infrastructure Services**

```bash
# Start MongoDB and HashiCorp Vault
docker-compose up -d

# Verify services are running
docker-compose ps
```

**Expected Output:**
```
NAME                 STATUS    PORTS
loyalcoin-mongodb    Up        0.0.0.0:27017->27017/tcp
loyalcoin-vault      Up        0.0.0.0:8200->8200/tcp
```

### **Step 3: Initialize Vault**

```bash
# Run Vault setup script
./setup_vault.sh
```

**This creates:**
- Transit encryption key for wallet security
- Root token: `dev-token` (saved to `.vault-token`)

### **Step 4: Configure Backend**

```bash
cd backend

# Create .env file from template
cp .env.example .env

# Edit .env file
nano .env
```

**Required Environment Variables:**

```bash
# Server
PORT=8080
HOST=0.0.0.0
ENV=development

# Database
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=loyalcoin_dev

# Cardano Blockchain
CARDANO_NETWORK=testnet
BLOCKFROST_PROJECT_ID=your_blockfrost_project_id_here
BLOCKFROST_API_URL=https://cardano-preprod.blockfrost.io/api/v0

# Policy & Token
LCN_POLICY_ID=
LCN_ASSET_NAME=4c434e
GOVERNANCE_WALLET_ADDRESS=

# Key Management (HashiCorp Vault)
VAULT_ADDR=http://localhost:8200
VAULT_TOKEN=dev-token
VAULT_TRANSIT_KEY=lcn-keys

# JWT Authentication
JWT_PRIVATE_KEY_PATH=./keys/jwt_private.pem
JWT_PUBLIC_KEY_PATH=./keys/jwt_public.pem
JWT_EXPIRATION_HOURS=24

# Security
BCRYPT_COST=12

# Transaction Settings
MIN_ADA_OUTPUT=1200000
FEE_A=155381
FEE_B=44
FEE_BUFFER_MULTIPLIER=1.2
CONFIRMATIONS_REQUIRED=3
WALLET_SEED_ADA=5000000

# Settlement
EXCHANGE_RATE_LCN_ETB=1.0
```

**⚠️ Important:** Replace `your_blockfrost_project_id_here` with your actual Blockfrost Project ID.

### **Step 5: Install Backend Dependencies**

```bash
# Download Go modules
go mod download

# Verify installation
go mod verify
```

### **Step 6: Install Node.js Dependencies for Scripts**

```bash
# Install transfer script dependencies
cd scripts/transfer
npm install
cd ../..
```

### **Step 7: Create Admin Account**

```bash
# Run admin creation utility
go run ./cmd/create-admin/main.go

# Follow prompts:
# Email: admin@loyalcoin.com
# Password: [create a strong password]
```

**This will:**
- Create admin user in MongoDB
- Generate Cardano wallet
- Display wallet address (save this!)

### **Step 8: Fund Admin Wallet**

```bash
# Copy the admin wallet address from step 7
# Visit Cardano Testnet Faucet
# https://docs.cardano.org/cardano-testnets/tools/faucet/

# Request 1000 test ADA to the admin wallet address
```

**Wait 1-2 minutes for transaction to confirm**, then verify:

```bash
# Check admin wallet balance
./check_admin_balance.sh
```

### **Step 9: Update .env with Admin Wallet**

```bash
# Edit .env file
nano .env

# Set GOVERNANCE_WALLET_ADDRESS to the admin wallet address
GOVERNANCE_WALLET_ADDRESS=addr_test1...
```

### **Step 10: Start Backend Server**

```bash
# Run backend
go run ./cmd/auth-service/main.go
```

**Expected Output:**
```
{"level":"info","msg":"Starting LoyalCoin Auth Service","env":"development"}
{"level":"info","msg":"Successfully connected to Vault"}
{"level":"info","msg":"Server starting","address":"0.0.0.0:8080"}
```

### **Step 11: Install & Start Frontends**

Open **3 separate terminal windows**:

#### **Terminal 1: Merchant Portal**
```bash
cd loyalcoin-merchant-portal
npm install
npm run dev
```
Access at: `http://localhost:5173`

#### **Terminal 2: Customer Portal**
```bash
cd loyalcoin-customer-portal
npm install
npm run dev
```
Access at: `http://localhost:5174`

#### **Terminal 3: Admin Portal**
```bash
cd loyalcoin-admin-portal
npm install
npm run dev
```
Access at: `http://localhost:5175`

---

## 🚀 **Production Deployment**

### **Backend Deployment (Render/Railway/Fly.io)**

#### **Step 1: Prepare Production Environment**

```bash
# Create production .env
cp .env.example .env.production

# Edit with production values
nano .env.production
```

**Production Environment Variables:**

```bash
ENV=production
PORT=8080
HOST=0.0.0.0

# Use MongoDB Atlas or managed MongoDB
MONGODB_URI=mongodb+srv://user:password@cluster.mongodb.net/
MONGODB_DATABASE=loyalcoin_prod

# Production Blockfrost Project
BLOCKFROST_PROJECT_ID=mainnet_project_id
BLOCKFROST_API_URL=https://cardano-mainnet.blockfrost.io/api/v0

# Production Vault (managed service)
VAULT_ADDR=https://vault.yourcompany.com
VAULT_TOKEN=production_vault_token
VAULT_TRANSIT_KEY=lcn-keys-prod

# Generate new JWT keys for production
JWT_PRIVATE_KEY_PATH=/app/keys/jwt_private.pem
JWT_PUBLIC_KEY_PATH=/app/keys/jwt_public.pem

# Production governance wallet
GOVERNANCE_WALLET_ADDRESS=addr1...
```

#### **Step 2: Create Dockerfile** (if deploying containerized)

```dockerfile
# Dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git nodejs npm

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Install Node.js script dependencies
WORKDIR /app/scripts/transfer
RUN npm install
WORKDIR /app/scripts/minting
RUN npm install

# Build backend
WORKDIR /app
RUN go build -o auth-service ./cmd/auth-service/main.go

# Production image
FROM alpine:latest

RUN apk --no-cache add ca-certificates nodejs npm

WORKDIR /root/

# Copy binary and scripts
COPY --from=builder /app/auth-service .
COPY --from=builder /app/scripts ./scripts

# Copy environment
COPY .env.production .env

EXPOSE 8080

CMD ["./auth-service"]
```

#### **Step 3: Deploy to Render**

1. **Create Render Account**: [render.com](https://render.com)

2. **Create New Web Service**:
   - Connect GitHub repository
   - Build Command: `go build -o server ./cmd/auth-service/main.go`
   - Start Command: `./server`
   - Environment: Go
   - Instance Type: Starter ($7/month)

3. **Add Environment Variables** in Render dashboard

4. **Enable Health Checks**:
   - Path: `/health`
   - Expected Status: 200

5. **Deploy** and wait for build to complete

**Note Backend URL**: `https://loyalcoin-api.onrender.com`

### **Frontend Deployment (Vercel/Netlify)**

#### **Deploy Admin Portal**

```bash
cd loyalcoin-admin-portal

# Install Vercel CLI
npm install -g vercel

# Login to Vercel
vercel login

# Deploy
vercel --prod
```

**Environment Variables on Vercel:**
```
VITE_API_URL=https://loyalcoin-api.onrender.com
```

**Repeat for:**
- Merchant Portal → `merchant.loyalcoin.io`
- Customer Portal → `app.loyalcoin.io`
- Admin Portal → `admin.loyalcoin.io`

---

## 🗄️ **Database Setup**

### **Local MongoDB (Docker)**

Already configured via `docker-compose.yml`.

### **Production MongoDB Atlas**

1. **Create Account**: [mongodb.com/cloud/atlas](https://www.mongodb.com/cloud/atlas)

2. **Create Cluster**:
   - Provider: AWS
   - Region: Closest to your backend (e.g., EU-WEST-1)
   - Tier: M10 or higher
   - Cloud Backup: Enabled

3. **Configure Access**:
   - Database Access → Add User → Set username/password
   - Network Access → Add IP → Allow from anywhere (0.0.0.0/0) *or whitelist backend IPs*

4. **Get Connection String**:
   ```
   mongodb+srv://username:password@cluster.abc.mongodb.net/loyalcoin_prod
   ```

5. **Add to Environment Variables**

---

## 🔐 **HashiCorp Vault Setup**

### **Production Vault**

**Option 1: HashiCorp Cloud Platform (HCP)**

1. Sign up: [cloud.hashicorp.com](https://cloud.hashicorp.com)
2. Create cluster: Vault Dedicated
3. Enable Transit Engine
4. Create service account with appropriate policies
5. Get Vault address and token

**Option 2: Self-Hosted Vault**

```bash
# On production server
docker run -d \
  --name vault \
  -p 8200:8200 \
  --cap-add=IPC_LOCK \
  -e 'VAULT_DEV_ROOT_TOKEN_ID=production-root-token' \
  hashicorp/vault:latest

# Initialize Vault
docker exec -it vault vault operator init

# Save unseal keys and root token securely!
```

---

## ⚙️ **System Configuration**

### **Nginx Reverse Proxy (Production)**

```nginx
# /etc/nginx/sites-available/loyalcoin-api

server {
    listen 80;
    server_name api.loyalcoin.io;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

**Enable HTTPS with Let's Encrypt:**

```bash
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d api.loyalcoin.io
```

### **Process Manager (PM2)**

```bash
# Install PM2
npm install -g pm2

# Create ecosystem file
cat > ecosystem.config.js <<EOF
module.exports = {
  apps: [{
    name: 'loyalcoin-backend',
    script: './auth-service',  
    cwd: '/opt/loyalcoin/backend',
    env: {
      NODE_ENV: 'production'
    },
    error_file: './logs/err.log',
    out_file: './logs/out.log',
    log_file: './logs/combined.log',
    time: true
  }]
}
EOF

# Start backend
pm2 start ecosystem.config.js

# Save PM2 configuration
pm2 save

# Setup PM2 to start on boot
pm2 startup systemd
```

---

## 📊 **Monitoring & Maintenance**

### **Health Checks**

```bash
# Backend health
curl http://localhost:8080/health

# MongoDB connection
mongosh mongodb://localhost:27017/loyalcoin_dev --eval "db.stats()"

# Vault status
docker exec vault vault status
```

### **Log Monitoring**

```bash
# Backend logs
tail -f backend/logs/combined.log

# MongoDB logs
docker logs -f loyalcoin-mongodb

# Vault logs
docker logs -f loyalcoin-vault
```

### **Database Backups**

```bash
# Create backup script
cat > /opt/scripts/backup-mongo.sh <<EOF
#!/bin/bash
DATE=\$(date +%Y%m%d_%H%M%S)
mongodump --uri="mongodb://localhost:27017/loyalcoin_prod" \
  --out="/backups/mongodb_\$DATE"
# Keep only last 7 days
find /backups -type d -mtime +7 -exec rm -rf {} \;
EOF

# Make executable
chmod +x /opt/scripts/backup-mongo.sh

# Setup daily cron job
crontab -e
# Add: 0 2 * * * /opt/scripts/backup-mongo.sh
```

---

## 🔍 **Troubleshooting**

### **Common Issues**

#### **Backend won't start - "address already in use"**

```bash
# Find process using port 8080
lsof -ti:8080

# Kill process
kill -9 $(lsof -ti:8080)

# Or use different port
PORT=8081 go run ./cmd/auth-service/main.go
```

#### **Vault authentication failed**

```bash
# Check Vault is running
docker ps | grep vault

# Get root token
docker logs loyalcoin-vault 2>&1 | grep "Root Token"

# Update .env with correct token
VAULT_TOKEN=dev-token
```

#### **"Insufficient funds" when transferring**

Check admin wallet has sufficient tADA:

```bash
# Get admin balance
curl http://localhost:8080/api/v1/admin/reserve/status \
  -H "Authorization: Bearer <admin_jwt_token>"
```

Fund wallet if needed from [testnet faucet](https://docs.cardano.org/cardano-testnets/tools/faucet/).

#### **Frontend can't connect to backend**

1. Check backend is running: `curl http://localhost:8080/health`
2. Verify `VITE_API_URL` in frontend `.env`
3. Check CORS settings in backend `middleware/cors.go`

#### **Transaction stuck as "PENDING"**

The transaction indexer will automatically update status. Wait 30-60 seconds, or manually trigger:

```bash
# Restart indexer (backend restart)
pkill -f auth-service
go run ./cmd/auth-service/main.go
```

---

## 🧪 **Testing**

### **Run API Tests**

```bash
# Full system test
./full_system_test.sh

# Individual tests
./test_signup.sh
./test_balance.sh  
./test_issue_lcn.sh
```

### **Integration Testing**

```bash
cd backend
go test ./... -v
```

---

## 📝 **Maintenance Checklist**

### **Weekly**
- [ ] Review error logs
- [ ] Check database size
- [ ] Verify backup completion
- [ ] Monitor Cardano testnet status

### **Monthly**
- [ ] Update dependencies (`go get -u`, `npm update`)
- [ ] Review security advisories
- [ ] Rotate JWT keys (if needed)
- [ ] Test disaster recovery procedures

### **Quarterly**
- [ ] Performance optimization review
- [ ] Database index optimization
- [ ] Vault audit log review
- [ ] Load testing

---

## 🆘 **Support**

**Documentation**: [docs.loyalcoin.io](https://docs.loyalcoin.io)  
**Email**: devops@loyalcoin.io  
**Discord**: [discord.gg/loyalcoin](https://discord.gg/loyalcoin)  

---

## 📚 **Additional Resources**

- [Cardano Documentation](https://docs.cardano.org)
- [Blockfrost API Docs](https://docs.blockfrost.io)
- [HashiCorp Vault Docs](https://developer.hashicorp.com/vault/docs)
- [MongoDB Atlas Guide](https://www.mongodb.com/docs/atlas/)
- [Vercel Deployment Guide](https://vercel.com/docs)

---

**Last Updated**: December 2025  
**Version**: 1.0.0
