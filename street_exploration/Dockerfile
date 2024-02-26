# Pull the base image with python 3.8 as a runtime for your Lambda
FROM public.ecr.aws/lambda/nodejs:20-x86_64

# Copy the earlier created index.mjs file to the container
COPY index.mjs ./

# Set the CMD to your handler
CMD ["index.handler"]