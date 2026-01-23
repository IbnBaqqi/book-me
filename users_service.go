package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// TODO look into timeout and context
func getUser42(accessToken string, url string) (*User42, error) {

	client := &http.Client{
		Timeout: time.Second * 15,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Request to get 42 user unsuccesful with status: %s", res.Status)
	}
	// close response body
	defer res.Body.Close()

	var user42 User42
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&user42)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	log.Println(user42)
	return &user42, nil
}
