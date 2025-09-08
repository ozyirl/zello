package internal

import (
	// "fmt"
	// "log"
	// "os"
	// "strconv"
	"fmt"
	"time"

	// "github.com/joho/godotenv"
	// "github.com/supabase-community/gotrue-go/types"
	"github.com/supabase-community/supabase-go"
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

func CreateIssue(client *supabase.Client,issueRequest CreateIssueRequest) (*Issue,error){
	var issues []Issue

	_,err := client.From("issues").Insert([]CreateIssueRequest{issueRequest},false,"","*","").ExecuteTo(&issues)
	if err != nil {
		return nil, fmt.Errorf("error while creating a issue %w",err)
	}


	if len(issues) == 0 {
		return nil, fmt.Errorf("insert succeeded but no issue returned")
	}

	return &issues[0], nil
	
}