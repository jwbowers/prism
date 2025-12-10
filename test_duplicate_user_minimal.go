package main

import (
	"fmt"
	"log"

	"github.com/scttfrdmn/prism/pkg/usermgmt"
)

func main() {
	// Create service
	storage := usermgmt.NewMemoryUserStorage()
	service := usermgmt.NewUserManagementService(storage)

	// Create first user
	user1 := &usermgmt.User{
		Username:    "testuser",
		Email:       "test1@example.com",
		DisplayName: "Test User 1",
	}

	fmt.Println("Creating first user...")
	err := service.CreateUser(user1)
	if err != nil {
		log.Fatalf("Failed to create first user: %v", err)
	}
	fmt.Printf("✅ First user created successfully with ID: %s\n", user1.ID)

	// List users to confirm
	users, _ := storage.ListUsers(nil)
	fmt.Printf("Total users in storage: %d\n", len(users))
	for _, u := range users {
		fmt.Printf("  - Username: %s, ID: %s, Email: %s\n", u.Username, u.ID, u.Email)
	}

	// Try to create second user with same username
	user2 := &usermgmt.User{
		Username:    "testuser", // Same username!
		Email:       "test2@example.com",
		DisplayName: "Test User 2",
	}

	fmt.Println("\nCreating second user with duplicate username...")
	err = service.CreateUser(user2)
	if err != nil {
		if err == usermgmt.ErrDuplicateUsername {
			fmt.Println("✅ SUCCESS: Correctly rejected duplicate username!")
		} else {
			log.Fatalf("Unexpected error: %v", err)
		}
	} else {
		fmt.Printf("❌ BUG CONFIRMED: Second user created with ID: %s (should have been rejected!)\n", user2.ID)

		// List users to see both
		users, _ := storage.ListUsers(nil)
		fmt.Printf("Total users in storage: %d (SHOULD BE 1, NOT 2!)\n", len(users))
		for _, u := range users {
			fmt.Printf("  - Username: %s, ID: %s, Email: %s\n", u.Username, u.ID, u.Email)
		}
	}
}
