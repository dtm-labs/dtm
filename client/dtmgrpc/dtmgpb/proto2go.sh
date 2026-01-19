# Step 1: Get and configure GOPATH (temporary effect)
# Get GOPATH and assign to variable (automatically adapt your path)
GOPATH=$(go env GOPATH)
# Add GOPATH/bin to PATH (solve plugin not found issue)
export PATH=$PATH:$GOPATH/bin

# Step 2: Verify if plugins exist (optional but recommended)
# Check if protoc-gen-go exists
if [ -f "$GOPATH/bin/protoc-gen-go" ]; then
  echo "protoc-gen-go plugin exists"
else
  echo "protoc-gen-go plugin doesn't exist, installing..."
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi

# Check if protoc-gen-go-grpc exists
if [ -f "$GOPATH/bin/protoc-gen-go-grpc" ]; then
  echo "protoc-gen-go-grpc plugin exists"
else
  echo "protoc-gen-go-grpc plugin doesn't exist, installing compatible version..."
  # Use version compatible with project gRPC v1.56.3
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0
fi


# Step 3: Generate Go code (core command)
# Generate basic pb.go file
protoc --go_out=. --go_opt=paths=source_relative dtmgimp.proto
# Generate grpc related pb.go file
protoc --go-grpc_out=. --go-grpc_opt=paths=source_relative dtmgimp.proto