#!/usr/bin/env bash
set -euo pipefail

PROJECT_ID="rally-go-6"
SERVICE_ACCOUNT="github-actions"
GITHUB_OWNER="Hoi-Trang-Huynh"
GITHUB_REPO="rally-backend-api"
POOL_NAME="github"
PROVIDER_NAME="github"
LOCATION="global"
PROTECTED_BRANCH="refs/heads/master"

echo "🔧 Bootstrapping GitHub OIDC for project: $PROJECT_ID"

PROJECT_NUMBER=$(gcloud projects describe "$PROJECT_ID" --format="value(projectNumber)")
SA_EMAIL="$SERVICE_ACCOUNT@$PROJECT_ID.iam.gserviceaccount.com"

# 1️⃣ Service account
if ! gcloud iam service-accounts describe "$SA_EMAIL" >/dev/null 2>&1; then
  echo "🆕 Creating service account $SA_EMAIL"
  gcloud iam service-accounts create "$SERVICE_ACCOUNT" \
    --project="$PROJECT_ID" \
    --display-name="GitHub Actions Deployer"
else
  echo "✔ Service account already exists"
fi

# 2️⃣ Roles
ROLES=(
  roles/run.admin
  roles/artifactregistry.writer
  roles/iam.serviceAccountUser
  roles/secretmanager.secretAccessor
)

for ROLE in "${ROLES[@]}"; do
  gcloud projects add-iam-policy-binding "$PROJECT_ID" \
    --member="serviceAccount:$SA_EMAIL" \
    --role="$ROLE" \
    --quiet
done

# 3️⃣ Pool
if ! gcloud iam workload-identity-pools describe "$POOL_NAME" \
  --project="$PROJECT_ID" \
  --location="$LOCATION" >/dev/null 2>&1; then

  echo "🆕 Creating workload identity pool"
  gcloud iam workload-identity-pools create "$POOL_NAME" \
    --project="$PROJECT_ID" \
    --location="$LOCATION" \
    --display-name="GitHub Actions Pool"
else
  echo "✔ Workload identity pool exists"
fi

# 4️⃣ Provider
if ! gcloud iam workload-identity-pools providers describe "$PROVIDER_NAME" \
  --project="$PROJECT_ID" \
  --location="$LOCATION" \
  --workload-identity-pool="$POOL_NAME" >/dev/null 2>&1; then

  echo "🆕 Creating workload identity provider"
  gcloud iam workload-identity-pools providers create-oidc "$PROVIDER_NAME" \
    --project="$PROJECT_ID" \
    --location="$LOCATION" \
    --workload-identity-pool="$POOL_NAME" \
    --display-name="GitHub Provider" \
    --issuer-uri="https://token.actions.githubusercontent.com" \
    --attribute-mapping="google.subject=assertion.sub,attribute.repository=assertion.repository,attribute.ref=assertion.ref,attribute.repository_owner=assertion.repository_owner" \
    --attribute-condition="assertion.repository_owner=='$GITHUB_OWNER' && assertion.repository=='$GITHUB_OWNER/$GITHUB_REPO' && (assertion.ref=='$PROTECTED_BRANCH' || assertion.ref.startsWith('refs/tags/'))"
else
  echo "🔄 Updating workload identity provider"
  gcloud iam workload-identity-pools providers update-oidc "$PROVIDER_NAME" \
    --project="$PROJECT_ID" \
    --location="$LOCATION" \
    --workload-identity-pool="$POOL_NAME" \
    --attribute-condition="assertion.repository_owner=='$GITHUB_OWNER' && assertion.repository=='$GITHUB_OWNER/$GITHUB_REPO' && (assertion.ref=='$PROTECTED_BRANCH' || assertion.ref.startsWith('refs/tags/'))"
fi

# 5️⃣ Bind repo
MEMBER="principalSet://iam.googleapis.com/projects/$PROJECT_NUMBER/locations/global/workloadIdentityPools/$POOL_NAME/attribute.repository/$GITHUB_OWNER/$GITHUB_REPO"

gcloud iam service-accounts add-iam-policy-binding "$SA_EMAIL" \
  --project="$PROJECT_ID" \
  --role="roles/iam.workloadIdentityUser" \
  --member="$MEMBER" \
  --quiet

# 6️⃣ Output
echo
echo "✅ GitHub Actions config:"
echo "workload_identity_provider: projects/$PROJECT_NUMBER/locations/global/workloadIdentityPools/$POOL_NAME/providers/$PROVIDER_NAME"
echo "service_account: $SA_EMAIL"
