# Dockerfile

# Use the official Go image as the base
FROM golang:latest

# Install air for hot-reloading
RUN go install github.com/air-verse/air@latest

# Set the working directory
WORKDIR /app

# Add the 'air' configuration file if you have one
COPY .air.toml /app/.air.toml

# Copy the source code into the container
COPY . /app

# Install ping utility
RUN apt-get update && apt-get install -y iputils-ping && apt-get install -y net-tools

# Default command to run air
CMD ["air"]

