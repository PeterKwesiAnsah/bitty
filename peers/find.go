package peers

import (
	"fmt"
	"io"
	"net/http"
)

// findPeers sends an HTTP GET request to the tracker with the specified parameters
// and returns the response body or an error if one occurred.
func FindPeers(trackerURL string) ([]byte, error) {
	// Send the HTTP GET request to the tracker
	resp, err := http.Get(trackerURL)
	if err != nil {
		return nil, fmt.Errorf("error announcing to tracker: %v", err)
	}

	defer resp.Body.Close()

	// Read and return the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}
	return body, nil
}
