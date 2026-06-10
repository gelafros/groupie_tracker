package internal

import (
	"log"
	"net/http"
)

// StartServer initializes the HTTP multiplexer, registers the routes,
// and starts the web server on port 8080.
func StartServer() {
	// Define the port the server will listen on
	port := ":8080"
	// Create a new HTTP multiplexer (router)
	handler := http.NewServeMux()

	// Print a message to the console indicating the server is starting
	log.Printf("Server is running on port: http://localhost%s\n", port)

	// Register the homeHandler function to handle requests to the root URL ("/")
	handler.HandleFunc("/", homeHandler)
	// Register the artistHandler to display individual artists
	handler.HandleFunc("/artist", artistHandler)
	// Handle the css call
	fileserver := http.FileServer(http.Dir("./static"))
	handler.Handle("/static/", http.StripPrefix("/static", fileserver))

	// Start the HTTP server; log.Fatal will catch and log any errors if the server fails to start
	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatal(err)
	}
}
