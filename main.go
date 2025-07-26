// main.go
// Simplified version that does not read the ErrorCode.
package main

import (
	"context"
	"log"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
)

func main() {
	endpoint := "opc.tcp://127.0.0.1:4840/freeopcua/server/"
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	log.SetFlags(0)

	c, err := opcua.NewClient(endpoint, opcua.SecurityMode(ua.MessageSecurityModeNone))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	if err := c.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer c.Close(ctx)
	log.Println("✅ Successfully connected to Robot Server!")
	time.Sleep(1 * time.Second)

	// --- Define NodeIDs directly ---
	ns := uint16(2) // Namespace index from server output
	robotNameNodeID := ua.NewNumericNodeID(ns, 1001)
	isActiveNodeID := ua.NewNumericNodeID(ns, 1002)
	speedNodeID := ua.NewNumericNodeID(ns, 1003)

	// --- Step 1: Read Initial Robot Status ---
	log.Println("--- Reading Initial Robot Status ---")
	readStringValue(ctx, c, robotNameNodeID, "RobotName")
	readBoolValue(ctx, c, isActiveNodeID, "IsActive")
	readFloatValue(ctx, c, speedNodeID, "Speed")

	// --- Step 2: Write Control Values ---
	log.Println("\n--- Sending Control Commands ---")
	setFloatValue(ctx, c, speedNodeID, "Speed", float32(2.5))
	setBoolValue(ctx, c, isActiveNodeID, "IsActive", true)

	log.Println("\n✅ Client finished.")
}

// Helper functions to read values.
func readStringValue(ctx context.Context, c *opcua.Client, nodeID *ua.NodeID, name string) {
	val, err := c.Node(nodeID).Value(ctx)
	if err != nil {
		log.Fatalf("❌ Failed to read %s: %v", name, err)
	}
	if val.Value() == nil {
		log.Printf("  - %s: (nil)", name)
		return
	}
	log.Printf("  - %s: '%s'", name, val.Value())
}

func readBoolValue(ctx context.Context, c *opcua.Client, nodeID *ua.NodeID, name string) {
	val, err := c.Node(nodeID).Value(ctx)
	if err != nil {
		log.Fatalf("❌ Failed to read %s: %v", name, err)
	}
	if val.Value() == nil {
		log.Printf("  - %s: (nil)", name)
		return
	}
	log.Printf("  - %s: %t", name, val.Value())
}

func readFloatValue(ctx context.Context, c *opcua.Client, nodeID *ua.NodeID, name string) {
	val, err := c.Node(nodeID).Value(ctx)
	if err != nil {
		log.Fatalf("❌ Failed to read %s: %v", name, err)
	}
	if val.Value() == nil {
		log.Printf("  - %s: (nil)", name)
		return
	}
	log.Printf("  - %s: %.2f", name, val.Value())
}

// Helper functions to write values.
func setFloatValue(ctx context.Context, c *opcua.Client, nodeID *ua.NodeID, name string, value float32) {
	variant, _ := ua.NewVariant(value)
	req := &ua.WriteRequest{
		NodesToWrite: []*ua.WriteValue{{NodeID: nodeID, AttributeID: ua.AttributeIDValue, Value: &ua.DataValue{Value: variant}}},
	}
	resp, err := c.Write(ctx, req)
	if err != nil {
		log.Printf("❌ Failed to write %s: %v", name, err)
	} else if resp.Results[0] != ua.StatusOK {
		log.Printf("❌ Failed to write %s with status %s", name, resp.Results[0])
	} else {
		log.Printf("  - Set %s to %.2f", name, value)
	}
}

func setBoolValue(ctx context.Context, c *opcua.Client, nodeID *ua.NodeID, name string, value bool) {
	variant, _ := ua.NewVariant(value)
	req := &ua.WriteRequest{
		NodesToWrite: []*ua.WriteValue{{NodeID: nodeID, AttributeID: ua.AttributeIDValue, Value: &ua.DataValue{Value: variant}}},
	}
	resp, err := c.Write(ctx, req)
	if err != nil {
		log.Printf("❌ Failed to write %s: %v", name, err)
	} else if resp.Results[0] != ua.StatusOK {
		log.Printf("❌ Failed to write %s with status %s", name, resp.Results[0])
	} else {
		log.Printf("  - Set %s to %t", name, value)
	}
}
