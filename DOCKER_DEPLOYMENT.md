# Docker Deployment Guide - GCP Instance

This guide walks you through deploying the AI Infrastructure Agent using Docker on a GCP Compute Engine instance.

## Prerequisites

- GCP Project with billing enabled
- gcloud CLI installed and authenticated
- Docker installed locally (for building images)
- Service account JSON key
- Ports: SSH (22), optional HTTP/HTTPS (80/443)

---

## Quick Start (Automated)

### Option 1: One-Command Deploy

```bash
cd ai-infra-agent
chmod +x deploy.sh
./deploy.sh YOUR_PROJECT_ID asia-south1 asia-south1-a
```

This script:
- ✓ Enables required GCP APIs
- ✓ Creates Artifact Registry repository
- ✓ Creates service account with proper roles
- ✓ Creates GCE instance
- ✓ Builds and pushes Docker image
- ✓ Deploys container to instance
- ✓ Sets up auto-restart

**Then wait 2-3 minutes and check logs:**
```bash
gcloud compute ssh ai-infra-agent --zone=asia-south1-a
docker logs -f ai-infra-agent
```

---

## Step-by-Step Manual Setup

### Step 1: Prepare GCP Environment

```bash
# Set project
export PROJECT_ID="your-gcp-project-id"
gcloud config set project $PROJECT_ID

# Enable APIs
gcloud services enable compute.googleapis.com
gcloud services enable aiplatform.googleapis.com
gcloud services enable artifactregistry.googleapis.com
```

### Step 2: Create Service Account

```bash
# Create service account
gcloud iam service-accounts create ai-infra-agent-sa \
    --display-name="AI Infrastructure Agent"

export SA_EMAIL="ai-infra-agent-sa@${PROJECT_ID}.iam.gserviceaccount.com"

# Grant roles
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member=serviceAccount:$SA_EMAIL \
    --role=roles/compute.admin

gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member=serviceAccount:$SA_EMAIL \
    --role=roles/aiplatform.user
```

### Step 3: Create Artifact Registry

```bash
export REGION="asia-south1"
gcloud artifacts repositories create ai-agents \
    --repository-format=docker \
    --location=$REGION \
    --description="AI Agent Docker images"
```

### Step 4: Create GCE Instance

```bash
export ZONE="asia-south1-a"
export INSTANCE_NAME="ai-infra-agent"

gcloud compute instances create $INSTANCE_NAME \
    --zone=$ZONE \
    --machine-type=e2-medium \
    --image-family=debian-11 \
    --image-project=debian-cloud \
    --service-account=$SA_EMAIL \
    --scopes=https://www.googleapis.com/auth/cloud-platform \
    --enable-display-device=false
```

### Step 5: Build and Push Docker Image

```bash
# Configure Docker auth
gcloud auth configure-docker $REGION-docker.pkg.dev

# Build image
export IMAGE_URL="$REGION-docker.pkg.dev/$PROJECT_ID/ai-agents/ai-infra-agent:latest"
docker build -t $IMAGE_URL .

# Push to Artifact Registry
docker push $IMAGE_URL
```

### Step 6: Deploy Container to Instance

```bash
# SSH into the instance
gcloud compute ssh $INSTANCE_NAME --zone=$ZONE

# On the instance, run these commands:
sudo apt-get update
sudo apt-get install -y docker.io
sudo systemctl start docker
sudo systemctl enable docker

# Configure Docker auth (uses service account credentials)
gcloud auth configure-docker asia-south1-docker.pkg.dev

# Create app directory
mkdir -p ~/ai-agent
cd ~/ai-agent

# Pull and run container
export IMAGE_URL="asia-south1-docker.pkg.dev/YOUR_PROJECT_ID/ai-agents/ai-infra-agent:latest"
sudo docker run -d \
    --name ai-infra-agent \
    --restart always \
    -e PROJECT_ID="YOUR_PROJECT_ID" \
    -e REGION="asia-south1" \
    -e ZONE="asia-south1-a" \
    -e GOOGLE_APPLICATION_CREDENTIALS=/app/sa-key.json \
    --log-driver json-file \
    --log-opt max-size=10m \
    --log-opt max-file=3 \
    $IMAGE_URL
```

### Step 7: Verify Deployment

```bash
# View running containers
sudo docker ps

# View logs
sudo docker logs -f ai-infra-agent

# Expected output:
# "Instance created: <instance-id>"
```

---

## Production Setup with Systemd

For better management, use a systemd service instead of docker run directly:

```bash
# Create Docker Compose file on instance
cat > ~/ai-agent/docker-compose.yml << 'EOF'
version: '3.8'
services:
  ai-infra-agent:
    image: asia-south1-docker.pkg.dev/YOUR_PROJECT_ID/ai-agents/ai-infra-agent:latest
    restart: always
    environment:
      PROJECT_ID: YOUR_PROJECT_ID
      REGION: asia-south1
      ZONE: asia-south1-a
      GOOGLE_APPLICATION_CREDENTIALS: /app/sa-key.json
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
EOF

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Create systemd service
sudo tee /etc/systemd/system/ai-infra-agent.service > /dev/null << 'EOF'
[Unit]
Description=AI Infrastructure Agent
After=docker.service
Requires=docker.service

[Service]
Type=simple
User=root
WorkingDirectory=/root/ai-agent
ExecStart=/usr/local/bin/docker-compose up
ExecStop=/usr/local/bin/docker-compose down
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable ai-infra-agent
sudo systemctl start ai-infra-agent
sudo systemctl status ai-infra-agent
```

---

## Monitoring & Logging

### View Container Logs

```bash
# On the instance
sudo docker logs -f ai-infra-agent

# Last 100 lines
sudo docker logs --tail 100 ai-infra-agent

# With timestamps
sudo docker logs -f --timestamps ai-infra-agent
```

### Cloud Logging

```bash
# View logs in GCP Console
gcloud logging read \
    'resource.type="gce_instance" AND resource.labels.instance_name="ai-infra-agent"' \
    --limit 50 \
    --format=json

# Real-time logs
gcloud logging read \
    'resource.type="gce_instance" AND resource.labels.instance_name="ai-infra-agent"' \
    --follow
```

### Monitor Resource Usage

```bash
# SSH into instance
gcloud compute ssh ai-infra-agent --zone=asia-south1-a

# Check container stats
sudo docker stats ai-infra-agent
```

---

## Troubleshooting

### Container won't start

```bash
sudo docker logs ai-infra-agent
# Check for:
# - Image pull errors → check Artifact Registry access
# - Environment variable issues → check -e flags
# - Port conflicts → check docker ps
```

### Authentication errors

```bash
# Verify service account has correct roles
gcloud projects get-iam-policy $PROJECT_ID \
    --flatten="bindings[].members" \
    --filter="bindings.members:serviceAccount:ai-infra-agent-sa*"

# Verify credentials on instance
gcloud auth list
gcloud config get-value project
```

### Image pull fails

```bash
# On instance, verify Docker auth
sudo docker pull $IMAGE_URL

# If fails, re-authenticate
gcloud auth login
gcloud auth configure-docker asia-south1-docker.pkg.dev
```

---

## Cleanup

```bash
# Delete instance
gcloud compute instances delete ai-infra-agent --zone=asia-south1-a

# Delete Artifact Registry
gcloud artifacts repositories delete ai-agents --location=asia-south1

# Delete service account
gcloud iam service-accounts delete ai-infra-agent-sa@${PROJECT_ID}.iam.gserviceaccount.com
```

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────┐
│              GCP Project                            │
├─────────────────────────────────────────────────────┤
│                                                     │
│  ┌───────────────────────────────────────────────┐  │
│  │  GCE Instance (ai-infra-agent)                │  │
│  │  Zone: asia-south1-a                          │  │
│  │  Type: e2-medium                              │  │
│  │                                               │  │
│  │  ┌───────────────────────────────────────┐   │  │
│  │  │ Docker Runtime                        │   │  │
│  │  │                                       │   │  │
│  │  │  ┌───────────────────────────────┐   │   │  │
│  │  │  │ ai-infra-agent container      │   │   │  │
│  │  │  │ - Go binary                   │   │   │  │
│  │  │  │ - Systemd restart policy      │   │   │  │
│  │  │  └───────────────────────────────┘   │   │  │
│  │  │                                       │   │  │
│  │  └───────────────────────────────────────┘   │  │
│  │                                               │  │
│  └───────────────────────────────────────────────┘  │
│           ↓                           ↓              │
│  ┌──────────────────┐      ┌──────────────────┐    │
│  │ Vertex AI API    │      │ Compute Engine   │    │
│  │ (Gemini 1.5)     │      │ API              │    │
│  └──────────────────┘      └──────────────────┘    │
│                                                     │
└─────────────────────────────────────────────────────┘
         ↓                            ↓
    Cloud Logging         Cloud Infrastructure
```

---

## Environment Variables

Create `.env` file for local docker-compose:

```bash
PROJECT_ID=your-gcp-project-id
REGION=asia-south1
ZONE=asia-south1-a
GOOGLE_APPLICATION_CREDENTIALS=/app/sa-key.json
```

---

## Security Best Practices

✓ Use service accounts (never hardcode credentials)  
✓ Apply least-privilege IAM roles  
✓ Use private GCE instances with Cloud NAT if possible  
✓ Enable VPC Flow Logs for network monitoring  
✓ Rotate service account keys regularly  
✓ Use secret management for sensitive configs  
✓ Enable audit logging on GCP resources

