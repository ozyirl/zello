package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"zel/lo/internal"

	"zel/lo/supabase"

	"github.com/joho/godotenv"
	authtypes "github.com/supabase-community/auth-go/types"
)




type CreateUserRequest struct {
	Name string `json:"name"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	supabaseUrl := os.Getenv("SUPABASE_URL")
	supabaseAnonKey := os.Getenv("SUPABASE_ANON_KEY")
	

	client, err := supabase.NewClient(supabaseUrl, supabaseAnonKey, &supabase.ClientOptions{})
	if err != nil {
		log.Fatal("Error creating Supabase client:", err)
	}
	var option string
	var userID string
	fmt.Println("press S to sign in and C to create account")
	fmt.Scanf("%s", &option)

	if option == "s"{
		var email string
		var password string
		fmt.Println("enter your email:")
		fmt.Scanf("%s", &email)
		fmt.Println("enter your password:")
		fmt.Scanf("%s", &password)
		session, err := client.SignInWithEmailPassword(email, password)
		if err != nil {
			log.Fatal("Error signing in:", err)
		}
		userID = session.User.ID.String()
		fmt.Println("Successfully authenticated!")
	} else if option == "c"{
		var email string
		var password string
		var name string
		fmt.Println("enter your email:")
		fmt.Scanf("%s", &email)
		fmt.Println("enter your password:")
		fmt.Scanf("%s", &password)
		fmt.Println("enter your name:")
		fmt.Scanf("%s", &name)
		
	
		signupRequest := authtypes.SignupRequest{
			Email:    email,
			Password: password,
		}
		_, err = client.Auth.Signup(signupRequest)
		if err != nil {
			log.Fatal("Error creating account:", err)
		}
		fmt.Println("Account created successfully!")
		
		
		session, err := client.SignInWithEmailPassword(email, password)
		if err != nil {
			log.Fatal("Error signing in:", err)
		}
		userID = session.User.ID.String()
		
	
		newUserRequest := CreateUserRequest{Name: name}
		user, err := CreateUser(client, newUserRequest)
		if err != nil {
			log.Fatal("Error creating user record:", err)
		}
		fmt.Printf("User record created: ID=%d, Name=%s\n", user.ID, user.Name)
	}

	// _, err = client.Auth.SignInWithEmailPassword(email, password)
	// if err != nil {
	// 	log.Fatal("Error signing in:", err)
	// }
	// fmt.Println("Successfully authenticated!")

	fmt.Println("press i to create an issue")
	fmt.Scanf("%s", &option)
	if option == "i"{
		var title string
		var description string
		fmt.Println("enter the title of the issue:")
		fmt.Scanf("%s", &title)
		fmt.Println("enter the description of the issue:")
		fmt.Scanf("%s", &description)
		issueRequest := internal.CreateIssueRequest{Title: title, Description: description}
		issue, err := internal.CreateIssue(client, issueRequest, userID)
		if err != nil {
			if strings.Contains(err.Error(), "EOF") {
				fmt.Println("Issue created (no return body due to RLS).")
			} else {
				log.Fatal("Error creating issue:", err)
			}
		} else {
			fmt.Printf("Issue created: ID=%d, Title=%s, Description=%s\n", issue.ID, issue.Title, issue.Description)
		}
	}



	



}





func CreateUser(client *supabase.Client, userRequest CreateUserRequest) (*internal.User, error) {
	var users []internal.User

	
	userData := map[string]interface{}{
		"name": userRequest.Name,
	}

	_, err := client.From("users").
		Insert([]map[string]interface{}{userData}, false, "", "*", "").
		ExecuteTo(&users)
	if err != nil {
		return nil, fmt.Errorf("failed to insert user: %w", err)
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("insert succeeded but no user returned")
	}

	return &users[0], nil
}
