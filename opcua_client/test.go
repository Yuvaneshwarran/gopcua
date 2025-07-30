package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
)

// Default and limit constants for OPC UA configuration
const (
	DefaultOpcuaResponseTimeout   = 5.0   // seconds
	DefaultOpcuaConnectionTimeout = 5.0   // seconds
	DefaultOpcuaDelay             = 500.0 // milliseconds

	MaxOpcuaResponseTimeout   = 600.0
	MaxOpcuaConnectionTimeout = 10.0
	MaxOpcuaDelay             = 5000.0
	MinOpcuaDelay             = 100.0
)

// OpcuaConfig holds the read/write operation details
type OpcuaConfig struct {
	Read  []interface{}
	Write []interface{}
}

// OpcuaRequest holds all parameters for a single task execution
type OpcuaRequest struct {
	EndpointURL       string
	SecurityPolicy    string
	SecurityMode      string
	AuthPolicy        string
	CertFile          string
	KeyFile           string
	ResponseTimeout   float64
	ConnectionTimeout float64
	Delay             float64
	Config            *OpcuaConfig
	client            *opcua.Client
	robot             string
}

// ProcessOpcua is the main entry point for handling an OPC UA task.
func ProcessOpcua(task map[string]interface{}, robot string) (bool, map[string]interface{}) {
	Infof("Processing OPC UA task: %v", task)
	var opcuaReq OpcuaRequest

	description, ok := task["description"].(map[string]interface{})
	if !ok {
		return false, map[string]interface{}{"status": false, "operation": "TASK_FAILURE", "message": "Missing or invalid 'description' in task"}
	}

	if err := validateOpcuaConfig(description); err != nil {
		return false, map[string]interface{}{"status": false, "operation": "TASK_FAILURE", "message": err.Error()}
	}

	opcuaReq.createOpcuaConfig(description, robot)

	client, err := opcuaReq.getOpcuaClient()
	if err != nil {
		Errorf("Unable to get OPC UA client: %v", err)
		return false, map[string]interface{}{"status": false, "operation": "TASK_FAILURE", "message": "Unable to establish a connection to OPC UA server"}
	}
	opcuaReq.client = client

	// This variable is now used below
	results := make([]map[string]interface{}, 0)

	// This variable is now used below
	config, _ := description["config"].(map[string]interface{})
	// This variable is now used below
	namespaceIndex, _ := extractIntField(config, "namespace_index")

	// --- THIS LOGIC WAS MISSING ---
	if len(opcuaReq.Config.Read) > 0 {
		for _, readConfig := range opcuaReq.Config.Read {
			data, err := opcuaReq.executeOpcuaRead(readConfig, namespaceIndex)
			if err != nil {
				// ... error handling ...
				Errorf("Error during OPC UA read: %v", err)
				return false, map[string]interface{}{"status": false, "operation": "TASK_FAILURE", "message": "Error reading data from OPC UA"}
			}
			result := map[string]interface{}{"data": map[string]interface{}{"value": data}, "status": true}
			results = append(results, result)
		}
		return true, map[string]interface{}{"status": true, "message": "OPC UA read operation successful", "results": results}
	}

	if len(opcuaReq.Config.Write) > 0 {
		for _, writeConfig := range opcuaReq.Config.Write {
			err := opcuaReq.executeOpcuaWrite(writeConfig, namespaceIndex)
			if err != nil {
				// ... error handling ...
				Errorf("Error during OPC UA write: %v", err)
				return false, map[string]interface{}{"status": false, "operation": "TASK_FAILURE", "message": "Error writing data from OPC UA"}
			}
			result := map[string]interface{}{"data": map[string]interface{}{"value": "OK"}, "status": true}
			results = append(results, result)
		}
		return true, map[string]interface{}{"status": true, "message": "OPC UA write operation successful", "results": results}
	}
	// --- END OF MISSING LOGIC ---

	return false, map[string]interface{}{"status": false, "operation": "TASK_INVALID", "message": "No OPC UA read or write operations specified"}
}

// createOpcuaConfig now receives the inner 'description' map.
func (opcuaReq *OpcuaRequest) createOpcuaConfig(description map[string]interface{}, robot string) {
	// --- CHANGE: Access connection and config from the description map ---
	connection, _ := description["connection"].(map[string]interface{})
	config, _ := description["config"].(map[string]interface{})

	opcuaReq.EndpointURL, _ = connection["endpoint_url"].(string)
	opcuaReq.SecurityPolicy, _ = connection["security_policy"].(string)
	opcuaReq.SecurityMode, _ = connection["security_mode"].(string)
	opcuaReq.AuthPolicy, _ = connection["auth_policy"].(string)
	opcuaReq.CertFile, _ = connection["cert_file"].(string)
	opcuaReq.KeyFile, _ = connection["key_file"].(string)

	opcuaReq.ResponseTimeout = ClampFloat(
		extractFloat64Field(connection, "response_timeout", DefaultOpcuaResponseTimeout),
		0.1, MaxOpcuaResponseTimeout,
	)
	opcuaReq.ConnectionTimeout = ClampFloat(
		extractFloat64Field(connection, "connection_timeout", DefaultOpcuaConnectionTimeout),
		0.1, MaxOpcuaConnectionTimeout,
	)
	opcuaReq.Delay = ClampFloat(
		extractFloat64Field(connection, "delay", DefaultOpcuaDelay),
		MinOpcuaDelay, MaxOpcuaDelay,
	)

	readConfig, _ := config["read"].([]interface{})
	writeConfig, _ := config["write"].([]interface{})

	opcuaReq.Config = &OpcuaConfig{Read: readConfig, Write: writeConfig}
	opcuaReq.robot = robot

	Infof("OPC UA config for robot %s: endpoint=%s, delay=%.0fms, connTimeout=%.1fs, respTimeout=%.1fs", robot, opcuaReq.EndpointURL, opcuaReq.Delay, opcuaReq.ConnectionTimeout, opcuaReq.ResponseTimeout)
}

// getOpcuaClient retrieves a cached client or creates a new one.
func (opcuaReq *OpcuaRequest) getOpcuaClient() (*opcua.Client, error) {
	key := opcuaReq.EndpointURL

	// Changed from global.OpcuaMu
	OpcuaMu.Lock()
	existingClient, exists := OpcuaClients[key]
	OpcuaMu.Unlock()

	if exists && existingClient != nil && existingClient.State() == opcua.Connected {
		return existingClient, nil
	}
	return opcuaReq.createNewOpcuaClient()
}

// createNewOpcuaClient establishes a new connection to the OPC UA server.
func (opcuaReq *OpcuaRequest) createNewOpcuaClient() (*opcua.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(opcuaReq.ConnectionTimeout)*time.Second)
	defer cancel()
	endpoints, err := opcua.GetEndpoints(ctx, opcuaReq.EndpointURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get endpoints: %w", err)
	}
	ep, err := opcua.SelectEndpoint(endpoints, opcuaReq.SecurityPolicy, ua.MessageSecurityModeFromString(opcuaReq.SecurityMode))
	if err != nil {
		return nil, fmt.Errorf("failed to select endpoint: %w", err)
	}
	opts := []opcua.Option{
		opcua.SecurityFromEndpoint(ep, ua.UserTokenTypeFromString(opcuaReq.AuthPolicy)),
		opcua.RequestTimeout(time.Duration(opcuaReq.ResponseTimeout) * time.Second),
	}
	if opcuaReq.CertFile != "" && opcuaReq.KeyFile != "" {
		opts = append(opts, opcua.CertificateFile(opcuaReq.CertFile), opcua.PrivateKeyFile(opcuaReq.KeyFile))
	}
	client, err := opcua.NewClient(opcuaReq.EndpointURL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OPC UA client: %w", err)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- client.Connect(ctx)
	}()

	select {
	case <-CancellationChannel[opcuaReq.robot]:
		Warnf("OPC UA connection cancelled for robot [%s]", opcuaReq.robot)
		client.Close(ctx)
		return nil, fmt.Errorf("connection to %s cancelled", opcuaReq.EndpointURL)
	case err := <-errCh:
		if err != nil {
			Errorf("Failed to connect to OPC UA server: %v", err)
			return nil, fmt.Errorf("unable to connect to the OPC UA server at %s", opcuaReq.EndpointURL)
		}
	case <-ctx.Done():
		return nil, fmt.Errorf("connection to OPC UA server timed out")
	}

	OpcuaMu.Lock()
	OpcuaClients[opcuaReq.EndpointURL] = client
	OpcuaMu.Unlock()

	return client, nil
}

// handleConnectionError manages the reconnection logic.
func (opcuaReq *OpcuaRequest) handleConnectionError(err error) bool {
	if opcuaReq.client != nil {
		opcuaReq.client.Close(context.Background())
	}
	Warnf("OPC UA connection error for robot [%s]: %v. Attempting to reconnect...", opcuaReq.robot, err)
	key := opcuaReq.EndpointURL
	ticker := time.NewTicker(time.Duration(opcuaReq.Delay) * time.Millisecond)
	defer ticker.Stop()
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Duration(opcuaReq.ConnectionTimeout)*time.Second)
	defer cancel()

	for {
		select {
		case <-CancellationChannel[opcuaReq.robot]:
			Errorf("Cancellation received, cancelling OPC UA reconnection")
			return false
		case <-InterruptChan:
			Errorf("Interrupt received, cancelling OPC UA reconnection")
			return false
		case <-timeoutCtx.Done():
			Errorf("Reconnection to %s timed out after %.1f seconds", key, opcuaReq.ConnectionTimeout)
			return false
		case <-ticker.C:
			Infof("Attempting to reconnect to %s...", key)
			newClient, connectErr := opcuaReq.createNewOpcuaClient()
			if connectErr == nil {
				Infof("Reconnected to %s successfully", key)
				opcuaReq.client = newClient
				return true
			}
			Errorf("Reconnect attempt to %s failed: %v", key, connectErr)
		}
	}
}

// executeOpcuaRead now takes a namespaceIndex to construct the NodeID.
func (opcuaReq *OpcuaRequest) executeOpcuaRead(readConfig interface{}, namespaceIndex int) (interface{}, error) {
	taskDetails, ok := readConfig.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid readConfig format")
	}

	nodeIDNum, err := extractIntField(taskDetails, "node_id")
	if err != nil {
		return nil, fmt.Errorf("invalid or missing 'node_id' field: %w", err)
	}

	// Construct the NodeID from the namespace and numeric ID
	nodeID := ua.NewNumericNodeID(uint16(namespaceIndex), uint32(nodeIDNum))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(opcuaReq.ResponseTimeout)*time.Second)
	defer cancel()

	v, err := opcuaReq.client.Node(nodeID).Value(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read value for node '%s': %w", nodeID.String(), err)
	}
	return v.Value(), nil
}

// executeOpcuaWrite now takes a namespaceIndex to construct the NodeID.
func (opcuaReq *OpcuaRequest) executeOpcuaWrite(writeConfig interface{}, namespaceIndex int) error {
	taskDetails, ok := writeConfig.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid type for writeConfig: expected map[string]interface{}")
	}

	nodeIDNum, err := extractIntField(taskDetails, "node_id")
	if err != nil {
		return fmt.Errorf("invalid or missing 'node_id' field: %w", err)
	}

	value, exists := taskDetails["value"]
	if !exists {
		return fmt.Errorf("missing 'value' field for write operation")
	}

	// Construct the NodeID from the namespace and numeric ID
	nodeID := ua.NewNumericNodeID(uint16(namespaceIndex), uint32(nodeIDNum))

	variant, err := ua.NewVariant(value)
	if err != nil {
		return fmt.Errorf("failed to create variant from value '%v': %w", value, err)
	}
	req := &ua.WriteRequest{
		NodesToWrite: []*ua.WriteValue{
			{
				NodeID:      nodeID,
				AttributeID: ua.AttributeIDValue,
				Value: &ua.DataValue{
					EncodingMask: ua.DataValueValue,
					Value:        variant,
				},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(opcuaReq.ResponseTimeout)*time.Second)
	defer cancel()

	_, err = opcuaReq.client.Write(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to write value to node '%s': %w", nodeID.String(), err)
	}
	return nil
}

// validateOpcuaConfig now receives the inner 'description' map.
func validateOpcuaConfig(description map[string]interface{}) error {
	// --- CHANGE: Look for keys inside the description map ---
	connection, ok := description["connection"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("missing or invalid 'connection' in description")
	}
	if _, ok := connection["endpoint_url"].(string); !ok {
		return fmt.Errorf("missing or invalid 'endpoint_url' in 'connection'")
	}
	config, ok := description["config"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("missing or invalid 'config' in description")
	}
	if _, err := extractIntField(config, "namespace_index"); err != nil {
		return fmt.Errorf("missing or invalid 'namespace_index' in 'config'")
	}
	_, readOk := config["read"].([]interface{})
	_, writeOk := config["write"].([]interface{})
	if !readOk && !writeOk {
		return fmt.Errorf("missing or invalid 'read' or 'write' in 'config'")
	}
	return nil
}

// --- Utility Functions (formerly in 'utils' package) ---

func ClampFloat(val, min, max float64) float64 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

func extractIntField(data map[string]interface{}, key string) (int, error) {
	raw, exists := data[key]
	if !exists {
		return 0, fmt.Errorf("missing field '%s'", key)
	}
	switch v := raw.(type) {
	case float64:
		return int(v), nil
	case int:
		return v, nil
	case string:
		parsed, err := strconv.Atoi(v)
		if err != nil {
			return 0, fmt.Errorf("field '%s' has invalid string value: %v", key, err)
		}
		return parsed, nil
	default:
		return 0, fmt.Errorf("unsupported type for field '%s': %T", key, v)
	}
}

func extractFloat64Field(data map[string]interface{}, key string, defaultVal float64) float64 {
	if val, ok := data[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case string:
			if parsed, err := strconv.ParseFloat(v, 64); err == nil {
				return parsed
			}
		}
	}
	return defaultVal
}
