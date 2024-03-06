# Kill container if running
docker kill pickup-selection || true

# Remove container if exists
docker rm pickup-selection || true
