# robot_server.py
# Simplified version without the ErrorCode simulation.

from opcua import Server, ua
import time
from datetime import datetime

# --- Setup the Server ---
server = Server()
server.set_endpoint("opc.tcp://0.0.0.0:4840/freeopcua/server/")
server.set_server_name("Industrial Robot Simulation Server")

uri = "http://examples.gopcua.com/robot"
idx = server.register_namespace(uri)

objects = server.get_objects_node()

# --- Define Robot with Numeric NodeIDs ---
robot_obj_id = ua.NodeId(1000, idx)
robot = objects.add_object(robot_obj_id, "MyRobot")

robot_name = robot.add_variable(ua.NodeId(1001, idx), "RobotName", "KUKA KR 210")
is_active = robot.add_variable(ua.NodeId(1002, idx), "IsActive", False)
speed = robot.add_variable(ua.NodeId(1003, idx), "Speed", 1.0)

# Make variables writable
is_active.set_writable(True)
speed.set_writable(True)

# --- Start the Server ---
server.start()
print(f"✅ Robot Simulation Server started at {server.endpoint}")
print(f"Namespace for the robot is index {idx}")

try:
    while True:
        # Safely get values before printing
        active_val = is_active.get_value()
        speed_val = speed.get_value()

        if active_val is None: active_val = False
        if speed_val is None: speed_val = 0.0

        now = datetime.now().strftime("%H:%M:%S")
        active_status = "Active" if active_val else "Inactive"
        print(f"[{now}] Status: {active_status}, Speed: {speed_val:.2f} m/s")

        time.sleep(2)

except KeyboardInterrupt:
    print("\n🛑 Shutting down server...")
finally:
    server.stop()
