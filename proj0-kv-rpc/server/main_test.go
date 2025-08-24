package main

import (
	"context"
	"log"
	"net"
	"testing"

	pb "kvrpc/model/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

// Create a dialer that connects to the buffconn listener
func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

// Create an in-memory gRPC server
func setupTestServer() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

// TestServer_SayHello_Direct tests the SayHello method directly
func TestServer_SayHello_Direct(t *testing.T) {
	// Create a server instance
	s := &server{}
	ctx := context.Background()

	// Test cases using table-driven testing
	tests := []struct {
		name           string
		request        *pb.HelloRequest
		expectedPrefix string
		expectError    bool
	}{
		{
			name:           "valid name",
			request:        &pb.HelloRequest{Name: "Alice"},
			expectedPrefix: "Hello Alice",
			expectError:    false,
		},
		{
			name:           "empty name",
			request:        &pb.HelloRequest{Name: ""},
			expectedPrefix: "Hello ",
			expectError:    false,
		},
		{
			name:           "name with spaces",
			request:        &pb.HelloRequest{Name: "John Doe"},
			expectedPrefix: "Hello John Doe",
			expectError:    false,
		},
		{
			name:           "name with special characters",
			request:        &pb.HelloRequest{Name: "José-María"},
			expectedPrefix: "Hello José-María",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the method directly
			response, err := s.SayHello(ctx, tt.request)

			// Check for errors
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify the response
			if response == nil {
				t.Errorf("Expected response but got nil")
				return
			}

			if response.Message != tt.expectedPrefix {
				t.Errorf("Expected message '%s', got '%s'", tt.expectedPrefix, response.Message)
			}
		})
	}
}

// TestServer_SayHello_Integration tests the full gRPC call flow
func TestServer_SayHello_Integration(t *testing.T) {
	// Setup the test server
	setupTestServer()

	// Create a client connection
	ctx := context.Background()
	var opts []grpc.DialOption = []grpc.DialOption{
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	conn, err := grpc.NewClient("passthrough://bufnet", opts...)
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	// Create the client
	client := pb.NewGreeterClient(conn)

	// Test cases
	tests := []struct {
		name        string
		requestName string
		expected    string
	}{
		{
			name:        "integration test with valid name",
			requestName: "Integration Test",
			expected:    "Hello Integration Test",
		},
		{
			name:        "integration test with empty name",
			requestName: "",
			expected:    "Hello ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make the gRPC call
			response, err := client.SayHello(ctx, &pb.HelloRequest{
				Name: tt.requestName,
			})

			if err != nil {
				t.Fatalf("SayHello failed: %v", err)
			}

			if response.Message != tt.expected {
				t.Errorf("Expected message '%s', got '%s'", tt.expected, response.Message)
			}
		})
	}
}

// TestServer_SayHello_NilRequest tests handling of nil requests
func TestServer_SayHello_NilRequest(t *testing.T) {
	s := &server{}
	ctx := context.Background()

	// This should not panic and should handle nil gracefully
	// Note: In practice, gRPC framework usually prevents nil requests,
	// but it's good to test defensive programming
	response, err := s.SayHello(ctx, nil)

	// The current implementation would panic on nil.GetName()
	// This test helps identify areas for improvement
	if err == nil && response != nil {
		t.Logf("Response for nil request: %s", response.Message)
	}
}

// BenchmarkServer_SayHello benchmarks the SayHello method
func BenchmarkServer_SayHello(b *testing.B) {
	s := &server{}
	ctx := context.Background()
	req := &pb.HelloRequest{Name: "Benchmark"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := s.SayHello(ctx, req)
		if err != nil {
			b.Fatalf("SayHello failed: %v", err)
		}
	}
}

// Example_sayHello demonstrates how to use the SayHello method
func Example_sayHello() {
	s := &server{}
	ctx := context.Background()
	req := &pb.HelloRequest{Name: "World"}

	response, err := s.SayHello(ctx, req)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	log.Printf("Response: %s", response.Message)
	// Output would be logged: Response: Hello World
}
