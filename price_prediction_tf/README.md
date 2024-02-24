# Uber Price Prediction (with TensorFlow)

## How to Install onto AWS

### Step 1. Log into AWS ECR

Run this command, replacing `<ecr_url>` with the ECR URL for the project (will be in Discord):
```
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin <ecr_url>
```

### Step 2a. Build the Docker container locally

This will take a long time when you build it for the first time, however subsequent builds should be much faster.
```
docker build -t price-prediction-tf .
```

### Step 2b. (optional) Run the Docker container locally and test it

Run this command to start the container:
```
docker run -p 8080:8080 price-prediction-tf
```

Next, open up your favorite REST API testing tool such as Postman.

To make a request to the API, the URL will be 
```
http://localhost:8080/2015-03-31/functions/function/invocations
```

and you will send a POST request with an application/json body containing the data.

### Step 3. Tag the Docker container for upload on ECR

```
docker tag price-prediction-tf <ecr_url>/pricing-lambda
```

### Step 4. Upload to ECR

```
docker push <ecr_url>/pricing-lambda
```

## Calling from AWS Lambda

This container expects (mile, minute) pairs in JSON form:

```json
{
    "miles": [
        0.5, 1.0, 2.0, 2.5, 4.0
    ],
    "minutes": [
        2.0, 4.0, 8.0, 23.5, 16.75
    ]
}
```

and returns data of the form:

```json
{
    "prices": [
        9.482263565063477,
        10.340690612792969,
        13.169022560119629,
        21.532272338867188,
        18.78050994873047
    ]
}
```