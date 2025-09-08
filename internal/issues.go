package internal

import (
	"fmt"
	"time"

	"zel/lo/supabase"
)


type Issue struct {
    ID          int       `json:"id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Status      string    `json:"status"`
    UserID      string    `json:"user_id"`   // UUID in Supabase
    CreatedAt   time.Time `json:"created_at"`
}

type CreateIssueRequest struct {
    Title       string `json:"title"`
    Description string `json:"description,omitempty"`
    Status      string `json:"status,omitempty"`
}


type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}


func CreateIssue(client *supabase.Client, issueRequest CreateIssueRequest, userID string) (*Issue, error) {
	var issues []Issue

	// Create the issue data with user ID
	issueData := map[string]interface{}{
		"title":       issueRequest.Title,
		"description": issueRequest.Description,
		"user_id":     userID,
	}

	// Set default status if not provided
	if issueRequest.Status != "" {
		issueData["status"] = issueRequest.Status
	} else {
		issueData["status"] = "open" // default status
	}

	_, err := client.From("issues").Insert([]map[string]interface{}{issueData}, false, "", "minimal", "").ExecuteTo(&issues)
	
	if err != nil {
		return nil, fmt.Errorf("error while creating a issue %w", err)
	}

	if len(issues) == 0 {
		return nil, fmt.Errorf("insert succeeded but no issue returned")
	}

	return &issues[0], nil
}