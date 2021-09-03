# Start from the latest golang base image
FROM golang:latest
# Add Maintainer Info
LABEL maintainer="Kashish <kashish@yellowmessenger.com>"
# Set the Current Working Directory inside the container
WORKDIR /app
# Copy go mod and sum files
COPY go.mod go.sum ./
# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download
# Copy the source from the current directory to the Working Directory inside the container
COPY . .
# Build the Go app
RUN go build -o asterisk-ari .
ENV NEW_RELIC_LICENSE_KEY=34ff2b9dd449161b5e0a5bbc161c06877355NRAL
RUN mkdir i -p /var/log/yellowmessenger/asterisk_ari/
RUN touch /var/log/yellowmessenger/asterisk_ari/asterisk_ari.log
# Expose port 9991 to the outside world
EXPOSE 9991
EXPOSE 8088
# Command to run the executable
ENTRYPOINT ["./asterisk-ari"]