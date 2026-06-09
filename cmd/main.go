package main

import "groupie-tracker/internal"

// main is the primary entry point for the groupie-tracker application.
func main() {
	// Call StartServer from the internal package to initialize and run the web application
	internal.StartServer()
}
