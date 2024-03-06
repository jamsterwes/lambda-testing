# Stop/remove previous container
./scripts/stop.sh

# Rebuild
docker build --platform linux/amd64 -t price-prediction-go .

# Run
docker run -d -p 8080:8080 --name price-prediction-go --platform linux/amd64 --entrypoint /usr/local/bin/aws-lambda-rie price-prediction-go ./main
