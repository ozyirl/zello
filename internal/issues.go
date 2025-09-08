package internal

import (
	"fmt"

	"zel/lo/supabase"
)


type Issue struct {
    ID          int       `json:"id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Status      string    `json:"status"`
    UserID      string    `json:"user_id"`   
   CreatedAt string `json:"created_at"`
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


func ListIssues(client *supabase.Client) ([]Issue, error) {
	var issues []Issue

	_, err := client.From("issues").Select("*", "", false).ExecuteTo(&issues)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch issues: %w", err)
	}

	return issues, nil
}



func CreateIssue(client *supabase.Client, issueRequest CreateIssueRequest, userID string) (*Issue, error) {
	var issues []Issue

	
	issueData := map[string]interface{}{
		"title":       issueRequest.Title,
		"description": issueRequest.Description,
		"user_id":     userID,
	}

	
	if issueRequest.Status != "" {
		issueData["status"] = issueRequest.Status
	} else {
		issueData["status"] = "open" 
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