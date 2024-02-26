# Street Exploration (soon to be Pickup Point Selector)

## Testing Locally

### Step 1. Build the Docker container locally

This will take a long time when you build it for the first time, however subsequent builds should be much faster.
```
docker build -t street-exploration .
```

### Step 2. Run the Docker container locally and test it

Run this command to start the container:
```
docker run -p 8080:8080 street-exploration
```

Next, open up your favorite REST API testing tool such as Postman.

To make a request to the API, the URL will be 
```
http://localhost:8080/2015-03-31/functions/function/invocations
```

and you will send a POST request with an application/json body containing the data.

## Uploading to AWS Lambda

For now, simply copy the contents of `index.mjs` into the Code Editor for this Lambda.

## Calling from AWS Lambda

This container expects the user's position:

```json
{
    "lat": 33.812975,
    "long": -88.134597
}
```

and returns an array of streets of the form:

```json
{
    "statusCode": 200,
    "body": [
        [
            {
                "lat": ...,
                "lon": ...
            },
            ...
        ],
        ...
    ]
}
```

This is a list of streets. Streets are stored as a list of points that are connected by line segments.