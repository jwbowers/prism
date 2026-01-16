package main

import (
	"fmt"

	"github.com/scttfrdmn/prism/pkg/usermgmt"
)

func main() {
	// Create storage and service
	storage := usermgmt.NewMemoryUserStorage()
	service := usermgmt.NewUserManagementService(storage)

	// Create first user
	user1 := &usermgmt.User{
		Username:    "testuser",
		Email:       "test1@example.com",
		DisplayName: "Test User 1",
	}

	fmt.Println("=== Creating first user ===")
	err := service.CreateUser(user1)
	if err != nil {
		fmt.Printf("❌ UNEXPECTED: Error creating first user: %v\n", err)
		return
	}
	fmt.Printf("✅ First user created: ID=%s, Username=%s\n", user1.ID, user1.Username)

	// Try to create duplicate user with same username
	user2 := &usermgmt.User{
		Username:    "testuser", // Same username!
		Email:       "test2@example.com",
		DisplayName: "Test User 2",
	}

	fmt.Println("\n=== Creating duplicate user ===")
	err = service.CreateUser(user2)
	if err == usermgmt.ErrDuplicateUsername {
		fmt.Println("✅ SUCCESS: Service correctly returned ErrDuplicateUsername")
	} else if err != nil {
		fmt.Printf("❌ FAIL: Service returned unexpected error: %v\n", err)
	} else {
		fmt.Printf("❌ FAIL: Service returned NO ERROR! Created user with ID=%s\n", user2.ID)
	}
}
