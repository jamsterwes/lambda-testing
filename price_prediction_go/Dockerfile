FROM golang:1.22.1 as build
WORKDIR /helloworld

# Copy dependencies list
COPY go.mod go.sum ./

# Download tensorflow libs
RUN wget https://storage.googleapis.com/tensorflow/libtensorflow/libtensorflow-cpu-linux-x86_64-2.15.0.tar.gz

# Copy over tensorflow lib
RUN mkdir ./tf
RUN tar -C ./tf -xzf libtensorflow-cpu-linux-x86_64-2.15.0.tar.gz
RUN tar -C /usr/local -xzf libtensorflow-cpu-linux-x86_64-2.15.0.tar.gz
RUN ldconfig

# Build with optional lambda.norpc tag
COPY *.go .
RUN go build -tags lambda.norpc -o main .

# Copy artifacts to a clean image
FROM public.ecr.aws/lambda/provided:al2023
COPY --from=build /helloworld/main ./main

# Copy the simple_model folder to the container
COPY tf2_model/ ./tf2_model

# Copy over tensorflow lib
COPY --from=build /helloworld/tf/lib ./

# Label main as entrypoint
ENTRYPOINT [ "./main" ]