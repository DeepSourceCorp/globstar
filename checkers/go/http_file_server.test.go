package main

import (
	"log"
	"net/http"
)

// Unsafe usage of http.FileServer
func unsafeFileServer() {
	// <expect-error> usage of http file server: serving files directly from an empty directory is unsafe
	http.Handle("/unsafe/", http.StripPrefix("/unsafe/", http.FileServer(http.Dir(""))))
}

// Safe usage with http.NewServeMux
func safeServer() {
	mux := http.NewServeMux()

	// Example of safe handler
	mux.HandleFunc("/safe", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("This is a safe handler using http.NewServeMux."))
	})

	log.Println("Safe server running at http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", mux))
}
