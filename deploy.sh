#!/bin/bash
# deploy.sh - Deploy AI Infra Agent to GCP with Docker

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ID=${1:-""}
REGION=${2:-"asia-south1"}
ZONE=${3:-"asia-south1-a"}
INSTANCE_NAME="ai-infra-agent"
MACHINE_TYPE="e2-medium"
IMAGE_FAMILY="debian-11"
IMAGE_PROJECT="debian-cloud"

if [ -z "$PROJECT_ID" ]; then
    echo -e "${RED}Error: PROJECT_ID is required${NC}"
    echo "Usage: ./deploy.sh <PROJECT_ID> [REGION] [ZONE]"
    exit 1
fi

echo -e "${YELLOW}=== AI Infra Agent Docker Deployment ===${NC}"
echo "Project ID: $PROJECT_ID"
echo "Region: $REGION"
echo "Zone: $ZONE"
echo ""

# Step 1: Set GCP project
echo -e "${YELLOW}Step 1: Setting GCP project...${NC}"
gcloud config set project $PROJECT_ID

# Step 2: Enable required APIs
echo -e "${YELLOW}Step 2: Enabling required APIs...${NC}"
gcloud services enable compute.googleapis.com
gcloud services enable aiplatform.googleapis.com
gcloud services enable artifactregistry.googleapis.com

# Step 3: Create Artifact Registry repository
echo -e "${YELLOW}Step 3: Creating Artifact Registry repository...${NC}"
REPO_NAME="ai-agents"
if ! gcloud artifacts repositories describe $REPO_NAME --location=$REGION &>/dev/null; then
    gcloud artifacts repositories create $REPO_NAME \
        --repository-format=docker \
        --location=$REGION \
        --description="AI Infrastructure Agent Docker images"
    echo -e "${GREEN}✓ Repository created${NC}"
else
    echo -e "${GREEN}✓ Repository already exists${NC}"
fi

# Step 4: Create service account
echo -e "${YELLOW}Step 4: Creating service account...${NC}"
SA_NAME="ai-infra-agent-sa"
if ! gcloud iam service-accounts describe $SA_NAME@$PROJECT_ID.iam.gserviceaccount.com &>/dev/null; then
    gcloud iam service-accounts create $SA_NAME \
        --display-name="AI Infrastructure Agent Service Account"
    echo -e "${GREEN}✓ Service account created${NC}"
else
    echo -e "${GREEN}✓ Service account already exists${NC}"
fi

SA_EMAIL="$SA_NAME@$PROJECT_ID.iam.gserviceaccount.com"

# Step 5: Grant roles to service account
echo -e "${YELLOW}Step 5: Granting IAM roles...${NC}"
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member=serviceAccount:$SA_EMAIL \
    --role=roles/compute.admin \
    --quiet

gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member=serviceAccount:$SA_EMAIL \
    --role=roles/aiplatform.user \
    --quiet

echo -e "${GREEN}✓ Roles assigned${NC}"

# Step 6: Create GCE instance
echo -e "${YELLOW}Step 6: Creating GCE instance...${NC}"
if gcloud compute instances describe $INSTANCE_NAME --zone=$ZONE &>/dev/null; then
    echo -e "${GREEN}✓ Instance already exists${NC}"
else
    gcloud compute instances create $INSTANCE_NAME \
        --zone=$ZONE \
        --machine-type=$MACHINE_TYPE \
        --image-family=$IMAGE_FAMILY \
        --image-project=$IMAGE_PROJECT \
        --service-account=$SA_EMAIL \
        --scopes=https://www.googleapis.com/auth/cloud-platform \
        --tags=http-server,https-server \
        --enable-display-device=false
    echo -e "${GREEN}✓ Instance created${NC}"
fi

# Step 7: Wait for instance to be ready
echo -e "${YELLOW}Step 7: Waiting for instance to be ready...${NC}"
sleep 10

# Step 8: Build and push Docker image
echo -e "${YELLOW}Step 8: Building and pushing Docker image...${NC}"
IMAGE_URL="$REGION-docker.pkg.dev/$PROJECT_ID/$REPO_NAME/$INSTANCE_NAME:latest"
echo "Image URL: $IMAGE_URL"

# Configure Docker authentication
gcloud auth configure-docker $REGION-docker.pkg.dev

# Build image
docker build -t $IMAGE_URL .

# Push image
docker push $IMAGE_URL
echo -e "${GREEN}✓ Image pushed${NC}"

# Step 9: Create startup script
echo -e "${YELLOW}Step 9: Creating startup script...${NC}"
cat > /tmp/startup.sh << 'EOF'
#!/bin/bash
set -e

# Update system
apt-get update
apt-get install -y docker.io

# Start Docker
systemctl start docker
systemctl enable docker

# Create app directory
mkdir -p /app
cd /app

# Configure Docker authentication (uses default service account)
gcloud auth configure-docker ${REGION}-docker.pkg.dev

# Pull and run container
docker pull ${IMAGE_URL}
docker run -d \
    --name ai-infra-agent \
    --restart always \
    -e PROJECT_ID=${PROJECT_ID} \
    -e REGION=${REGION} \
    -e ZONE=${ZONE} \
    -e GOOGLE_APPLICATION_CREDENTIALS=/app/sa-key.json \
    --log-driver json-file \
    --log-opt max-size=10m \
    --log-opt max-file=3 \
    ${IMAGE_URL}

# Log startup
docker logs -f ai-infra-agent
EOF

# Step 10: Deploy startup script to instance
echo -e "${YELLOW}Step 10: Deploying to instance...${NC}"
gcloud compute instances add-metadata $INSTANCE_NAME \
    --zone=$ZONE \
    --metadata-from-file startup-script=/tmp/startup.sh

echo -e "${YELLOW}Step 11: Restarting instance to run startup script...${NC}"
gcloud compute instances stop $INSTANCE_NAME --zone=$ZONE --async
sleep 5
gcloud compute instances start $INSTANCE_NAME --zone=$ZONE --async

echo -e "${GREEN}✓ Deployment initiated${NC}"
echo ""
echo -e "${GREEN}=== Deployment Complete ===${NC}"
echo ""
echo "Next steps:"
echo "1. Wait 2-3 minutes for the instance to start and container to deploy"
echo "2. SSH into the instance:"
echo "   gcloud compute ssh $INSTANCE_NAME --zone=$ZONE"
echo "3. Check container logs:"
echo "   docker logs -f ai-infra-agent"
echo ""
echo "To monitor from Cloud Logging:"
echo "   gcloud logging read 'resource.type=gce_instance AND resource.labels.instance_id=${INSTANCE_NAME}' --limit 50"
