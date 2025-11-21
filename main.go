package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
	"io/ioutil"
)

var startTime time.Time

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

func loadSwaggerSpec(filePath string) (string, error) {
	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("file not found at. Err: %s", err)
		return "", err
	}

	// Read the entire file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Printf("error reading file. Err: %s", err)
		return "", err
	}

	return string(data), nil
}

func swagger(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Server", "http-swagger-server")
	const swaggerPath = "/config/swagger.json"
	swaggerJSON, err := ioutil.ReadFile(swaggerPath)
	if err != nil {
		http.Error(w, "Could not read swagger.json", http.StatusInternalServerError)
		log.Printf("Error reading swagger.json: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(swaggerJSON)
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
	http.HandleFunc(swaggerEndpoint, swagger)

	log.Println("Version 0.3.0")

	errs := Run(httpPort)

	// This will run forever until channel receives error
	select {
	case err := <-errs:
		log.Printf("Could not start serving service due to (error: %s)", err)
	}
}
