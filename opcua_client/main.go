package main

import (
	"fmt"
)

func main() {
	// --- Example Usage with New Structure ---
	// You must have an OPC UA server that has nodes with these numeric IDs.

	// Example 1: Read tasks with description and numeric node IDs.
	readTask := map[string]interface{}{
		"description": "Read initial state of the robot sensors.",
		"connection": map[string]interface{}{
			"endpoint_url": "opc.tcp://127.0.0.1:4840",
		},
		"config": map[string]interface{}{
			"namespace_index": 2, // Specify the namespace for all nodes in this config
			"read": []interface{}{
				map[string]interface{}{
					"node_id": 5001, // Simplified numeric NodeID for Temperature
				},
				map[string]interface{}{
					"node_id": 5002, // Simplified numeric NodeID for Status
				},
			},
		},
	}

	fmt.Println("--- 1. Executing Read Task ---")
	success, result := ProcessOpcua(readTask, "robot1")
	fmt.Printf("Success: %v\nResult: %v\n\n", success, result)

	// Example 2: Write task with new structure.
	writeTask := map[string]interface{}{
		"description": "Set robot status to 'Processing'.",
		"connection": map[string]interface{}{
			"endpoint_url": "opc.tcp://127.0.0.1:4840",
		},
		"config": map[string]interface{}{
			"namespace_index": 2,
			"write": []interface{}{
				map[string]interface{}{
					"node_id": 5002,
					"value":   "Processing Task 1",
				},
			},
		},
	}

	fmt.Println("--- 2. Executing Write Task ---")
	success, result = ProcessOpcua(writeTask, "robot1")
	fmt.Printf("Success: %v\nResult: %v\n\n", success, result)

	// Example 3: Verify the write.
	verifyTask := map[string]interface{}{
		"description": "Verify the new robot status.",
		"connection": map[string]interface{}{
			"endpoint_url": "opc.tcp://127.0.0.1:4840",
		},
		"config": map[string]interface{}{
			"namespace_index": 2,
			"read": []interface{}{
				map[string]interface{}{
					"node_id": 5002,
				},
			},
		},
	}
	fmt.Println("--- 3. Verifying Write by Reading Back ---")
	success, result = ProcessOpcua(verifyTask, "robot1")
	fmt.Printf("Success: %v\nResult: %v\n", success, result)
}
