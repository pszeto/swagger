package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
	"io/ioutil"
	"io"
)

var startTime time.Time

// EchoResponse defines the structure for the JSON response
type EchoResponse struct {
	Headers map[string][]string `json:"headers"`
	Path    string              `json:"path"`
	Arguments    map[string]string `json:"arguments"`
	Method       string            `json:"method"`
	Origin      string              `json:"origin"`
	URL         string              `json:"url"`
	Body    string              `json:"body"`
	EnvVars    map[string]string   `json:"env_vars"`
}

func uptime() time.Duration {
	return time.Since(startTime)
}

func init() {
	startTime = time.Now()
}

func status(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Server", "http-swagger-server")
	resp := make(map[string]string)
	resp["uptime"] = uptime().String()
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error happened in JSON marshal. Err: %s", err)
	}
	w.Write(jsonResp)
	return
}

func swagger(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Server", "http-swagger-server")
	const swaggerPath = "/config/swagger.json"
	swaggerJSON, err := ioutil.ReadFile(swaggerPath)
	if err != nil {
		log.Printf("Error reading swagger.json: %v", err)
		log.Println("Loading default swagger from filesystem")
	}
	swaggerJSON, err = ioutil.ReadFile("./swagger.json")
	w.Header().Set("Content-Type", "application/json")
	w.Write(swaggerJSON)
}

// splitEnvVar splits an environment variable string into key and value.
func splitEnvVar(env string) []string {
	for i := 0; i < len(env); i++ {
		if env[i] == '=' {
			return []string{env[:i], env[i+1:]}
		}
	}
	return []string{env}
}

func echo(w http.ResponseWriter, r *http.Request) {
	// Read the request body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	bodyString := string(bodyBytes)

	// Parse query arguments
	queryParams := make(map[string]string)
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			queryParams[key] = values[0] // Take the first value if multiple are present
		}
	}

	// Get environment variables
	envVars := make(map[string]string)
	for _, env := range os.Environ() {
		parts := splitEnvVar(env)
		if len(parts) == 2 {
			envVars[parts[0]] = parts[1]
		}
	}
	// Create the response object
	response := EchoResponse{
		Headers: r.Header, // r.Header is already a map[string][]string
		Path: 	 r.URL.Path,
		Arguments:    queryParams,
		Method:       r.Method,
		Origin:      r.RemoteAddr,
		URL:         r.URL.String(),
		Body:    bodyString,
		EnvVars: envVars,
	}

	// Set the Content-Type header to application/json
	w.Header().Set("Content-Type", "application/json")

	// Encode the response object to JSON and write it to the response writer
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
		return
	}
}

func Run(addr string) chan error {

	errs := make(chan error)

	// Starting HTTP server
	go func() {
		log.Printf("Staring HTTP service on %s", addr)

		if err := http.ListenAndServe(addr, nil); err != nil {
			errs <- err
		}

	}()

	return errs
}

func main() {
	httpPort, ok := os.LookupEnv("HTTP_PORT")
	if !ok {
		log.Println("HTTP_PORT not defined.  Defaulting to 8080")
		httpPort = ":8080"
	} else {
		httpPort = ":" + httpPort
	}

	swaggerEndpoint, ok := os.LookupEnv("SWAGGER_ENDPOINT")
	if !ok {
		log.Println("SWAGGER_ENDPOINT not defined.  Defaulting to /swagger.json")
		swaggerEndpoint = "/swagger.json"
	}

	http.HandleFunc("/status", status)
	http.HandleFunc("/", echo)
	http.HandleFunc(swaggerEndpoint, swagger)

	log.Println("Version 0.4.0")

	errs := Run(httpPort)

	// This will run forever until channel receives error
	select {
	case err := <-errs:
		log.Printf("Could not start serving service due to (error: %s)", err)
	}
}
