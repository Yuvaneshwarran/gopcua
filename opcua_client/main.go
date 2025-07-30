package main

import (
	"fmt"
)

func main() {
	// Example 1: Read task with the new nested structure.
	readTask := map[string]interface{}{
		"description": map[string]interface{}{
			"connection": map[string]interface{}{
				"endpoint_url": "opc.tcp://127.0.0.1:4840",
			},
			"config": map[string]interface{}{
				"namespace_index": 2,
				"read": []interface{}{
					map[string]interface{}{
						"node_id": 5001, // Numeric NodeID for Temperature
					},
					map[string]interface{}{
						"node_id": 5002, // Numeric NodeID for Status
					},
				},
			},
		},
	}

	fmt.Println("--- 1. Executing Read Task ---")
	success, result := ProcessOpcua(readTask, "robot1")
	fmt.Printf("Success: %v\nResult: %v\n\n", success, result)

	// Example 2: Write task with the new nested structure.
	writeTask := map[string]interface{}{
		"description": map[string]interface{}{
			"connection": map[string]interface{}{
				"endpoint_url": "opc.tcp://127.0.0.1:4840",
			},
			"config": map[string]interface{}{
				"namespace_index": 2,
				"write": []interface{}{
					map[string]interface{}{
						"node_id": 5002,
						"value":   "Processing Task 2",
					},
				},
			},
		},
	}

	fmt.Println("--- 2. Executing Write Task ---")
	success, result = ProcessOpcua(writeTask, "robot1")
	fmt.Printf("Success: %v\nResult: %v\n\n", success, result)

	// Example 3: Read task with the new nested structure.
	readTask = map[string]interface{}{
		"description": map[string]interface{}{
			"connection": map[string]interface{}{
				"endpoint_url": "opc.tcp://127.0.0.1:4840",
			},
			"config": map[string]interface{}{
				"namespace_index": 2,
				"read": []interface{}{
					map[string]interface{}{
						"node_id": 5001, // Numeric NodeID for Temperature
					},
					map[string]interface{}{
						"node_id": 5002, // Numeric NodeID for Status
					},
				},
			},
		},
	}

	fmt.Println("--- 3. Executing Read Task ---")
	success, result = ProcessOpcua(readTask, "robot1")
	fmt.Printf("Success: %v\nResult: %v\n\n", success, result)
}
