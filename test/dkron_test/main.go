package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	DkronServer = "http://localhost:8080"
	JobCount    = 3000
)

type Job struct {
	Name       string `json:"name"`
	Schedule   string `json:"schedule"`
	Command    string `json:"command"`
	Owner      string `json:"owner"`
	OwnerEmail string `json:"owner_email"`
	Disabled   bool   `json:"disabled"`
}

func main() {
	for i := 1; i <= JobCount; i++ {
		job := Job{
			Name:       fmt.Sprintf("job-%d", i),
			Schedule:   "@every 1s",
			Command:    "echo hello",
			Owner:      "tester",
			OwnerEmail: "tester@example.com",
			Disabled:   false,
		}

		jobJSON, err := json.Marshal(job)
		if err != nil {
			fmt.Printf("Error marshalling job %d: %s\n", i, err)
			continue
		}

		resp, err := http.Post(DkronServer+"/v1/jobs", "application/json", bytes.NewBuffer(jobJSON))
		if err != nil {
			fmt.Printf("Error creating job %d: %s\n", i, err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Failed to create job %d. Status: %s\n", i, resp.Status)
		}

		resp.Body.Close()
		time.Sleep(10 * time.Millisecond)
	}

	fmt.Println("Finished creating jobs.")
}
