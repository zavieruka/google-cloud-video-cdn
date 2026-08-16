#!/bin/bash
set -e

echo "========================================="
echo "Video Platform - Processed Bucket Delivery Setup"
echo "========================================="
echo ""

PROJECT_ID=$(gcloud config get-value project)

if [ -z "$PROJECT_ID" ]; then
    echo "Error: No GCP project configured. Please run:"
    echo "  gcloud config set project YOUR_PROJECT_ID"
    exit 1
fi

# The processed bucket and the runtime service account. Override via env if your
# naming differs from the defaults documented in the README.
PROCESSED_BUCKET_NAME="${PROCESSED_BUCKET_NAME:-${PROJECT_ID}-videos-processed}"
SERVICE_ACCOUNT_EMAIL="${SERVICE_ACCOUNT_EMAIL:-video-platform-dev@${PROJECT_ID}.iam.gserviceaccount.com}"

# Browser origins allowed to fetch HLS segments and upload custom thumbnails
# directly to Cloud Storage.
# Comma-separated; "*" allows any. Keep this in sync with CORS_ALLOWED_ORIGINS.
CORS_ALLOWED_ORIGINS="${CORS_ALLOWED_ORIGINS:-*}"

echo "Project ID:        $PROJECT_ID"
echo "Processed bucket:  gs://$PROCESSED_BUCKET_NAME"
echo "Service account:   $SERVICE_ACCOUNT_EMAIL"
echo "Allowed origins:   $CORS_ALLOWED_ORIGINS"
echo ""

# Build a JSON array of origins from the comma-separated list.
ORIGINS_JSON=$(printf '%s' "$CORS_ALLOWED_ORIGINS" | awk -F',' '{
    out=""
    for (i = 1; i <= NF; i++) {
        gsub(/^[ \t]+|[ \t]+$/, "", $i)
        if ($i == "") continue
        if (out != "") out = out ", "
        out = out "\"" $i "\""
    }
    print out
}')

CORS_FILE=$(mktemp)
trap 'rm -f "$CORS_FILE"' EXIT

# HLS players issue byte-range GET/HEAD requests for fMP4 segments, while custom
# thumbnails use signed PUT uploads. Expose range-related headers for seeking.
cat > "$CORS_FILE" <<EOF
[
  {
    "origin": [${ORIGINS_JSON}],
    "method": ["GET", "HEAD", "PUT"],
    "responseHeader": ["Content-Type", "Content-Length", "Content-Range", "Range", "Accept-Ranges", "ETag"],
    "maxAgeSeconds": 3600
  }
]
EOF

echo "Applying CORS configuration to the processed bucket..."
gcloud storage buckets update "gs://${PROCESSED_BUCKET_NAME}" --cors-file="$CORS_FILE"

echo ""
echo "Granting the service account object access on the processed bucket..."
echo "(Delivery is via signed URLs; the bucket is NOT made public.)"
gcloud storage buckets add-iam-policy-binding "gs://${PROCESSED_BUCKET_NAME}" \
    --member="serviceAccount:${SERVICE_ACCOUNT_EMAIL}" \
    --role="roles/storage.objectUser"

echo ""
echo "========================================="
echo "Processed Bucket Delivery Setup Complete!"
echo "========================================="
echo ""
echo "Configured:"
echo "  ✓ CORS [GET, HEAD, PUT] for: ${CORS_ALLOWED_ORIGINS}"
echo "  ✓ Service account object access (objectUser)"
echo "  ✓ Bucket remains private — no allUsers binding"
echo ""
echo "Note: signing download URLs also requires the runtime identity to hold"
echo "      roles/iam.serviceAccountTokenCreator on ${SERVICE_ACCOUNT_EMAIL}"
echo "      (the same permission used for signed upload URLs)."
echo ""
