# Kill container if running
docker kill price-prediction-go || true

# Remove container if exists
docker rm price-prediction-go || true
