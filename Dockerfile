# ---- Build Stage ----

# this specifies the base image
# as tells what we will call it in the rest of the file
# also it is beneficial for multi stage builds as if we don't name our stages we have to use integers
# which depends on the number of from statements starting from integer 0
# this is a issue because the order might change and file can break if the necessary changes are not made
FROM golang:1.24-alpine AS builder 

# working directory inside the container
WORKDIR /app

# this below instructions tells the builder to copy 
# the go.mod and go.sum files to the working directory
COPY go.mod go.sum ./

#run this specific command [builder will run this]
RUN go mod download

# copy <source> <destination> - 
# here we copy everything from current directory on host to working directory in the image
# before this we need to set the work directory as if we don't
# destination will be /* wjicj means we are polluting the root directory.
# if we need to exclude anything add to dockerignore file
COPY . .

#CGO stands for call C libraries , CGO_ENABLED=0 means compile pure go only
#as the dependencies required for c libraries are not present ,

# GOOS=linux means compile thinking the target OS is Linux
# docker runs on linux and without this it might compile some other binary
# which linux container cannot compile

# GOARCH=amd64 means compile for 64 bit cpu ,
# this is done to gurantee compaitability
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -o mini-redis ./cmd/server

# ---- Runtime Stage ----
# distroles image is basically an image containing the program and its runtime
# dependencies and not the shells, package managers etc
#hence these are very small compared to using alpine directly [ 2mb vs 5mb ]
FROM gcr.io/distroless/base-debian12

WORKDIR /app
# By default, the COPY instruction copies files from the build context. 
# The COPY --from flag lets you copy files from an image, a build stage, 
# or a named context instead.
# here we are copying from our build stage
COPY --from=builder /app/mini-redis .

#this is the port we will export
EXPOSE 6379
ENTRYPOINT ["/app/mini-redis"]
# the entrypoint here is defined in vector form here instead of
# normally because distroless doesn't have a shell hence using
# ENTRYPOINT "/app/mini-redis" will give error