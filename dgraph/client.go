package dgraph

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/dgraph-io/dgo/v230"
	"github.com/dgraph-io/dgo/v230/protos/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client represents a Dgraph client with connection management
type Client struct {
	dgraphClient *dgo.Dgraph
	conn         *grpc.ClientConn
	config       *Config
}

// Config holds Dgraph connection configuration
type Config struct {
	Host           string        `json:"host"`
	Port           string        `json:"port"`
	MaxRetries     int           `json:"max_retries"`
	RetryDelay     time.Duration `json:"retry_delay"`
	RequestTimeout time.Duration `json:"request_timeout"`
}

// QueryResponse represents the response from Dgraph query execution
type QueryResponse struct {
	Data      interface{} `json:"data"`
	QueryTime string      `json:"query_time"`
	Success   bool        `json:"success"`
	Error     string      `json:"error,omitempty"`
}

// ExecutionStats represents query execution statistics
type ExecutionStats struct {
	QueryTime    time.Duration `json:"query_time"`
	ResultCount  int           `json:"result_count"`
	TotalQueries int           `json:"total_queries"`
	CacheHit     bool          `json:"cache_hit"`
	ExecutedAt   time.Time     `json:"executed_at"`
}

// DefaultConfig returns default Dgraph client configuration
func DefaultConfig() *Config {
	return &Config{
		Host:           "localhost",
		Port:           "9080",
		MaxRetries:     3,
		RetryDelay:     time.Second * 2,
		RequestTimeout: time.Second * 30,
	}
}

// NewClient creates a new Dgraph client with the given configuration
func NewClient(config *Config) (*Client, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Create gRPC connection
	conn, err := grpc.Dial(
		fmt.Sprintf("%s:%s", config.Host, config.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial Dgraph: %w", err)
	}

	// Create Dgraph client
	dgraphClient := dgo.NewDgraphClient(api.NewDgraphClient(conn))

	client := &Client{
		dgraphClient: dgraphClient,
		conn:         conn,
		config:       config,
	}

	// Test connection
	if err := client.TestConnection(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("connection test failed: %w", err)
	}

	log.Printf("✅ Connected to Dgraph at %s:%s", config.Host, config.Port)
	return client, nil
}

// TestConnection tests the connection to Dgraph
func (c *Client) TestConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), c.config.RequestTimeout)
	defer cancel()

	// Simple health check query
	query := `{ health(func: has(dgraph.type)) { count(uid) } }`

	_, err := c.dgraphClient.NewTxn().Query(ctx, query)
	if err != nil {
		return fmt.Errorf("health check query failed: %w", err)
	}

	return nil
}

// ExecuteDQL executes a DQL query and returns the results
func (c *Client) ExecuteDQL(ctx context.Context, query string) (*QueryResponse, error) {
	start := time.Now()

	// Set timeout if not already set in context
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.config.RequestTimeout)
		defer cancel()
	}

	// Execute query with retries
	var response *api.Response
	var err error

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		response, err = c.dgraphClient.NewTxn().Query(ctx, query)
		if err == nil {
			break
		}

		if attempt < c.config.MaxRetries {
			log.Printf("⚠️ Query attempt %d failed, retrying in %v: %v",
				attempt+1, c.config.RetryDelay, err)
			time.Sleep(c.config.RetryDelay)
		}
	}

	queryTime := time.Since(start)

	if err != nil {
		return &QueryResponse{
			Data:      nil,
			QueryTime: queryTime.String(),
			Success:   false,
			Error:     err.Error(),
		}, err
	}

	// Parse JSON response
	var data interface{}
	if len(response.Json) > 0 {
		if err := json.Unmarshal(response.Json, &data); err != nil {
			return &QueryResponse{
				Data:      nil,
				QueryTime: queryTime.String(),
				Success:   false,
				Error:     fmt.Sprintf("failed to parse response JSON: %v", err),
			}, err
		}
	}

	return &QueryResponse{
		Data:      data,
		QueryTime: queryTime.String(),
		Success:   true,
	}, nil
}

// ExecuteMultipleDQL executes multiple DQL queries and returns combined results
func (c *Client) ExecuteMultipleDQL(ctx context.Context, queries []string) (map[string]*QueryResponse, error) {
	results := make(map[string]*QueryResponse)

	for i, query := range queries {
		queryName := fmt.Sprintf("query_%d", i+1)

		result, err := c.ExecuteDQL(ctx, query)
		if err != nil {
			// Continue with other queries even if one fails
			log.Printf("⚠️ Query %s failed: %v", queryName, err)
		}

		results[queryName] = result
	}

	return results, nil
}

// GetExecutionStats returns statistics about query execution
func (c *Client) GetExecutionStats(response *QueryResponse) *ExecutionStats {
	stats := &ExecutionStats{
		ExecutedAt: time.Now(),
		CacheHit:   false, // DQL queries are not cached by default
	}

	// Parse query time
	if queryTime, err := time.ParseDuration(response.QueryTime); err == nil {
		stats.QueryTime = queryTime
	}

	// Count results
	if response.Data != nil {
		if dataMap, ok := response.Data.(map[string]interface{}); ok {
			for _, value := range dataMap {
				if array, ok := value.([]interface{}); ok {
					stats.ResultCount += len(array)
					stats.TotalQueries++
				}
			}
		}
	}

	return stats
}

// Close closes the connection to Dgraph
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// IsConnected checks if the client is connected to Dgraph
func (c *Client) IsConnected() bool {
	if c.dgraphClient == nil || c.conn == nil {
		return false
	}

	// Quick health check with short timeout
	return c.TestConnection() == nil
}
