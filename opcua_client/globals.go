package main

import (
	"log"
	"os"
	"sync"

	"github.com/gopcua/opcua"
)

// OpcuaMu protects access to the OpcuaClients map.
var OpcuaMu sync.Mutex

// OpcuaClients caches active OPC UA client connections.
var OpcuaClients = make(map[string]*opcua.Client)

// CancellationChannel allows for external cancellation of tasks for a specific robot.
var CancellationChannel = make(map[string]chan struct{})

// InterruptChan is a global channel to signal an application-wide shutdown.
var InterruptChan = make(chan os.Signal, 1)

// SimpleLogger provides a basic logging implementation.
var SimpleLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

// We can add simple wrapper functions to match the original code's logging style.
func Infof(format string, v ...interface{}) {
	SimpleLogger.Printf(format, v...)
}

func Warnf(format string, v ...interface{}) {
	SimpleLogger.Printf("WARN: "+format, v...)
}

func Errorf(format string, v ...interface{}) {
	SimpleLogger.Printf("ERROR: "+format, v...)
}

// The Writer() function has been removed as it is not needed and causes a compile error.
