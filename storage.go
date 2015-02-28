/*
 *
 *
 *
 */
package main

import "path/filepath"

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

// The size of the copy buffer from the http to the file system.
const BUFFER_SIZE = 8 * 1024

// Handler of the StorageProtocol
//
// Providing the functionality for access of PUT, GET, POST, DELETE to
// arbitrary path
type StorageHandler struct {

	// root directory for storing and retrieve the files
	RootDir string
}

// Split a file path into the dirname and basename
//
//
func PathSplit(path string) (string, string) {
	return filepath.Dir(path), filepath.Base(path)
}

func ServeGet(w http.ResponseWriter, r *http.Request, path string) {
	// Copying file into the response
	file, err := os.Open(path)

	if err == nil {
		w.WriteHeader(200)
		buffer := make([]byte, BUFFER_SIZE)

		for {
			count, err := file.Read(buffer)

			w.Write(buffer[:count])

			if err != nil {
				break
			}
		}

		if err != io.EOF {
			fmt.Printf("error during reading file: %s \n", err)
		}
		fmt.Fprintf(w, "ok")
	}
}

func ServePut(w http.ResponseWriter, r *http.Request, path string) {
	dir, _ := PathSplit(path)
	os.MkdirAll(dir, 0755)

	buffer, err := ioutil.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "error %s", err)
		return
	}

	err = ioutil.WriteFile(path, buffer, 0755)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "error %s", err)
		return
	}

	w.WriteHeader(200)
	fmt.Fprintf(w, "ok")

}

func ServePost(w http.ResponseWriter, r *http.Request, path string) {
	w.WriteHeader(501)
	fmt.Fprintf(w, "not implemented")
}

func ServeDelete(w http.ResponseWriter, r *http.Request, path string) {
	err := os.RemoveAll(path)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "err: %s", err)
	}
}

// Handling the FileStorage Protocol
func (sh *StorageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	uri, _ := filepath.Abs(r.RequestURI)
	method := r.Method
	path := sh.RootDir + uri

	fmt.Printf("Request to: %s as %s\n", uri, method)
	fmt.Printf("Routing to: %s", path)

	switch method {
	case "GET":
		ServeGet(w, r, path)
	case "PUT":
		ServePut(w, r, path)
	case "DELETE":
		ServeDelete(w, r, path)
	case "POST":
		ServePost(w, r, path)
	}
}

// retrieve an option from the environment,
// if no such value exists, fallback is return
func get_option(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func main() {
	bind_host := get_option("STORAGE_HOST", ":8080")
	fmt.Printf("Starting server, listening on %s\n", bind_host)
	//	http.HandleFunc("/", handler)
	handler := new(StorageHandler)

	handler.RootDir = get_option("STORAGE_ROOT", "./test")
	err := http.ListenAndServe(":8080", handler)

	if err != nil {
		fmt.Println(err)
	}
}
