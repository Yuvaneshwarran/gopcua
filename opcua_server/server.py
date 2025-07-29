import time
from opcua import Server

# --- 1. Setup the Server ---
server = Server()

# Set the endpoint URL
server.set_endpoint("opc.tcp://0.0.0.0:4840/freeopcua/server/")

# Set the server name
server.set_server_name("My Python OPC UA Server")

# --- 2. Create Custom Namespace and Nodes ---

# Register a new namespace for our custom nodes
uri = "http://examples.gopcua.org/robot"
idx = server.register_namespace(uri)

# Get the Objects folder, which is the standard place for custom nodes
objects = server.get_objects_node()

# Create a "Robot" object folder within our namespace to organize variables
robot = objects.add_object(idx, "Robot")

# Add the "Robot.Temperature" variable with a string NodeID
# The NodeID will be ns=2;s=Robot.Temperature
temp_var = robot.add_variable(f"ns={idx};s=Robot.Temperature", "Robot.Temperature", 25.0)
temp_var.set_writable() # Allow clients to change this value

# Add the "Robot.Status" variable with a string NodeID
# The NodeID will be ns=2;s=Robot.Status
status_var = robot.add_variable(f"ns={idx};s=Robot.Status", "Robot.Status", "Idle")
status_var.set_writable() # Allow clients to change this value

# --- 3. Start the Server and Run ---
server.start()
print(f"Server started at {server.endpoint}")
print(f"Custom namespace index is {idx}")
print("Nodes available at:")
print(f"  - {temp_var.nodeid}")
print(f"  - {status_var.nodeid}")

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