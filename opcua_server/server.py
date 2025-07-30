import time
from opcua import Server, ua

# --- 1. Setup the Server ---
server = Server()
server.set_endpoint("opc.tcp://0.0.0.0:4840/freeopcua/server/") # Set the endpoint URL
server.set_server_name("My Python OPC UA Server") # Set the server name

# --- 2. Create Custom Namespace and Nodes ---
uri = "http://examples.gopcua.org/robot" # Register a new namespace for our custom nodes
idx = server.register_namespace(uri)
objects = server.get_objects_node() # Get the Objects folder, which is the standard place for custom nodes
robot = objects.add_object(idx, "Robot") # Create a "Robot" object folder within our namespace to organize variables

# Create node ID for robot_temperature
temp_node_id = ua.NodeId(5001, idx)
temp_var = robot.add_variable(temp_node_id, "Robot.Temperature", 25.0)
temp_var.set_writable() # Allow clients to change this value

# Create node ID for robot_status
status_node_id = ua.NodeId(5002, idx)
status_var = robot.add_variable(status_node_id, "Robot.Status", "Idle")
status_var.set_writable() # Allow clients to change this value 

# --- 3. Start the Server and Run ---
server.start()
print(f"Server started at {server.endpoint}")
print(f"Custom namespace index is {idx}")
print("Nodes available at:")
print(f"  - {temp_var.nodeid}") # should now print ns=2; i=5001
print(f"  - {status_var.nodeid}") # should now print ns=2; i=5002

try:
    # Keep the server running and update temperature every 2 seconds
    count = 0
    while True:
        time.sleep(2)
        count += 0.1
        temp_var.set_value(25.0 + count)

except KeyboardInterrupt:
    print("\nShutting down server...")
finally:
    # Close the server
    server.stop()
    print("Server stopped.")