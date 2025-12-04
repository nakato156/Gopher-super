FROM golang:1.23-alpine

WORKDIR /app

# Install git for go mod download if needed, though not strictly required if vendored
# RUN apk add --no-cache git

# Copy go.mod and go.sum from the root to handle dependencies correctly
COPY go.mod go.sum ./
RUN go mod download

# Copy the specific directory for this service
COPY movie_lens_data_procc ./movie_lens_data_procc

# Set working directory to where the scripts are
WORKDIR /app/movie_lens_data_procc

# Build the applications (optional, can also just run go run)
# But running go run is fine for scripts
# We will use a shell script or chained command for the entrypoint

CMD sh -c "go run analisis.go && go run upload_to_mongodb.go -uri mongodb://mongo:27017 && go run upload_movies.go -uri mongodb://mongo:27017"
