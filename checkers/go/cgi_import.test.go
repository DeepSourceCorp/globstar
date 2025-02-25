package main

import (
	// <expect-error> usage of cgi package
	"net/http/cgi"
	"fmt"
	"net/http"
)

func cgi_test() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello from CGI (Insecure and Deprecated)!")
	})

	// Using the deprecated cgi package to serve HTTP requests (not recommended)
	err := cgi.Serve(handler)
	if err != nil {
		fmt.Printf("Error running CGI server: %v\n", err)
	}
}

func http_test() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello from the secure HTTP server!")
	})

	// Secure and modern way to serve HTTP requests
	fmt.Println("Starting secure server on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}
