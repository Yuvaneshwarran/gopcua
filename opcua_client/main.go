package main

import (
	"fmt"
)

func main() {
	// --- Example Usage ---
	// You must have an OPC UA server running at this endpoint.

	// Example 1: Read the initial Temperature.
	readTask := map[string]interface{}{
		"connection": map[string]interface{}{
			"endpoint_url": "opc.tcp://127.0.0.1:4840", // Using 127.0.0.1
		},
		"config": map[string]interface{}{
			"read": []interface{}{
				map[string]interface{}{
					"node_id": "ns=2;s=Robot.Temperature",
				},
			},
		},
	}

	fmt.Println("--- 1. Executing Initial Read Task ---")
	success, result := ProcessOpcua(readTask, "robot1")
	fmt.Printf("Success: %v\nResult: %v\n\n", success, result)

	// Example 2: Write a new Status.
	writeTask := map[string]interface{}{
		"connection": map[string]interface{}{
			"endpoint_url": "opc.tcp://127.0.0.1:4840",
		},
		"config": map[string]interface{}{
			"write": []interface{}{
				map[string]interface{}{
					"node_id": "ns=2;s=Robot.Status",
					"value":   "Processing Task 1",
				},
			},
		},
	}

	fmt.Println("--- 2. Executing Write Task ---")
	success, result = ProcessOpcua(writeTask, "robot1")
	fmt.Printf("Success: %v\nResult: %v\n\n", success, result)

	// =========================================================
	// === ADDED: Step 3 to verify the write operation       ===
	// =========================================================
	readBackTask := map[string]interface{}{
		"connection": map[string]interface{}{
			"endpoint_url": "opc.tcp://127.0.0.1:4840",
		},
		"config": map[string]interface{}{
			"read": []interface{}{
				map[string]interface{}{
					// Reading the same node we just wrote to
					"node_id": "ns=2;s=Robot.Status",
				},
			},
		},
	}

	fmt.Println("--- 3. Verifying Write by Reading Back ---")
	success, result = ProcessOpcua(readBackTask, "robot1")
	fmt.Printf("Success: %v\nResult: %v\n", success, result)
}
