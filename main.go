package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/supabase-community/gotrue-go/types"
	"github.com/supabase-community/supabase-go"
)


type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}


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
	fmt.Println("press S to sign in and C to create account")
	fmt.Scanf("%s", &option)

	if option == "s"{
		var email string
		var password string
		fmt.Println("enter your email:")
		fmt.Scanf("%s", &email)
		fmt.Println("enter your password:")
		fmt.Scanf("%s", &password)
		_, err = client.Auth.SignInWithEmailPassword(email, password)
	if err != nil {
		log.Fatal("Error signing in:", err)
	}
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
		
	
		signupRequest := types.SignupRequest{
			Email:    email,
			Password: password,
		}
		_, err = client.Auth.Signup(signupRequest)
		if err != nil {
			log.Fatal("Error creating account:", err)
		}
		fmt.Println("Account created successfully!")
		
		
		_, err = client.Auth.SignInWithEmailPassword(email, password)
		if err != nil {
			log.Fatal("Error signing in:", err)
		}
		
	
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

	var choice string
	fmt.Println("Enter I for insertion, G to get user info, or L to list all users:")
	fmt.Scanf("%s", &choice)

	if choice == "i" {
		fmt.Println("You chose insertion")
		var name string
		fmt.Println("enter the username")
		fmt.Scan(&name)

		newUserRequest := CreateUserRequest{Name: name}

		user, err := CreateUser(client, newUserRequest)
		if err != nil {
			log.Fatal("error creating user:", err)
		}
		fmt.Printf("Created user: ID=%d, Name=%s\n", user.ID, user.Name)

	} else if choice == "g" {
		var id int
		fmt.Println("You chose get user info")
		fmt.Println("enter the user id you want")
		fmt.Scanf("%d", &id)

		user, err := fetchUser(client, id)
		if err != nil {
			log.Fatal("Error fetching user:", err)
		}

		fmt.Printf("Fetched user: ID=%d, Name=%s\n", user.ID, user.Name)

	} else if choice == "l" {
		fmt.Println("You chose list all users")
		users, err := listAllUsers(client)
		if err != nil {
			log.Fatal("Error listing users:", err)
		}

		if len(users) == 0 {
			fmt.Println("No users found in the database")
		} else {
			fmt.Printf("Found %d users:\n", len(users))
			for _, user := range users {
				fmt.Printf("  ID=%d, Name=%s\n", user.ID, user.Name)
			}
		}

	} else {
		fmt.Println("Invalid choice")
	}



	



}







func fetchUser(client *supabase.Client, userID int) (*User, error) {
	var users []User
	
	_, err := client.From("users").Select("*", "", false).Eq("id", strconv.Itoa(userID)).ExecuteTo(&users)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("user with ID %d not found", userID)
	}

	return &users[0], nil
}


func listAllUsers(client *supabase.Client) ([]User, error) {
	var users []User
	
	_, err := client.From("users").Select("*", "", false).ExecuteTo(&users)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}
	
	return users, nil
}

func CreateUser(client *supabase.Client, userRequest CreateUserRequest) (*User, error) {
	var users []User

	
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
