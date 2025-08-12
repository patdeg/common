package main

import (
	"context"
	"fmt"
	"log"

	"github.com/patdeg/common"
	"github.com/patdeg/common/datastore"
	"github.com/patdeg/common/email"
)

// Example user struct
type User struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Age   int    `json:"age"`
}

func main() {
	fmt.Println("=== Common Package Basic Usage Example ===")

	// 1. Basic logging with PII protection
	demonstrateLogging()

	// 2. Type conversion utilities
	demonstrateConversions()

	// 3. Slice operations
	demonstrateSliceOperations()

	// 4. Data storage (uses local storage in development)
	demonstrateDataStorage()

	// 5. Email service (uses local mode in development)
	demonstrateEmailService()

	fmt.Println("\n=== Example completed successfully! ===")
}

func demonstrateLogging() {
	fmt.Println("\n1. Logging Examples:")
	
	// Regular logging
	common.Info("Application started")
	common.Warn("This is a warning message")
	
	// PII-safe logging (automatically masks sensitive data)
	userEmail := "john.doe@example.com"
	common.InfoSafe("User logged in: %s", userEmail)
	
	// Debug logging (only shows in development)
	common.Debug("Debug information: processing user data")
	
	fmt.Println("   ✓ Logging examples completed")
}

func demonstrateConversions() {
	fmt.Println("\n2. Type Conversion Examples:")
	
	// String to int conversion
	if num, err := common.StringToInt("123"); err == nil {
		fmt.Printf("   String '123' converted to int: %d\n", num)
	}
	
	// Int to string conversion
	str := common.IntToString(456)
	fmt.Printf("   Int 456 converted to string: '%s'\n", str)
	
	// Bool conversion
	if b, err := common.StringToBool("true"); err == nil {
		fmt.Printf("   String 'true' converted to bool: %t\n", b)
	}
	
	fmt.Println("   ✓ Type conversion examples completed")
}

func demonstrateSliceOperations() {
	fmt.Println("\n3. Slice Operations Examples:")
	
	numbers := []int{1, 2, 3, 2, 4, 3, 5}
	
	// Check if slice contains element
	contains := common.Contains(numbers, 3)
	fmt.Printf("   Slice contains 3: %t\n", contains)
	
	// Remove duplicates
	unique := common.Unique(numbers)
	fmt.Printf("   Original: %v\n", numbers)
	fmt.Printf("   Unique:   %v\n", unique)
	
	// Filter elements
	evens := common.Filter(numbers, func(n int) bool { return n%2 == 0 })
	fmt.Printf("   Even numbers: %v\n", evens)
	
	// Map/transform elements
	doubled := common.Map(numbers, func(n int) int { return n * 2 })
	fmt.Printf("   Doubled: %v\n", doubled)
	
	fmt.Println("   ✓ Slice operations examples completed")
}

func demonstrateDataStorage() {
	fmt.Println("\n4. Data Storage Examples:")
	
	ctx := context.Background()
	
	// Initialize repository (automatically uses local storage in development)
	repo, err := datastore.NewRepository(ctx)
	if err != nil {
		log.Printf("   Error initializing repository: %v", err)
		return
	}
	
	// Create a user
	user := &User{
		Email: "john.doe@example.com",
		Name:  "John Doe",
		Age:   30,
	}
	
	// Save user
	err = repo.Put(ctx, "User", user.Email, user)
	if err != nil {
		log.Printf("   Error saving user: %v", err)
		return
	}
	fmt.Printf("   ✓ Saved user: %s\n", user.Email)
	
	// Retrieve user
	var retrieved User
	err = repo.Get(ctx, "User", user.Email, &retrieved)
	if err != nil {
		log.Printf("   Error retrieving user: %v", err)
		return
	}
	fmt.Printf("   ✓ Retrieved user: %s (age: %d)\n", retrieved.Name, retrieved.Age)
	
	// Clean up (delete user)
	err = repo.Delete(ctx, "User", user.Email)
	if err != nil {
		log.Printf("   Error deleting user: %v", err)
	} else {
		fmt.Printf("   ✓ Deleted user: %s\n", user.Email)
	}
	
	fmt.Println("   ✓ Data storage examples completed")
}

func demonstrateEmailService() {
	fmt.Println("\n5. Email Service Examples:")
	
	// Initialize email service (uses local mode in development)
	service, err := email.NewService(email.Config{
		Provider:  "local", // Will print emails to console instead of sending
		FromEmail: "noreply@example.com",
		FromName:  "Example App",
	})
	if err != nil {
		log.Printf("   Error initializing email service: %v", err)
		return
	}
	
	// Create and send a welcome email
	message := &email.Message{
		To: []email.Address{
			{Email: "user@example.com", Name: "New User"},
		},
		Subject: "Welcome to our service!",
		HTML:    `<h1>Welcome!</h1><p>Thank you for joining our service.</p>`,
		Text:    "Welcome!\n\nThank you for joining our service.",
	}
	
	err = service.Send(context.Background(), message)
	if err != nil {
		log.Printf("   Error sending email: %v", err)
		return
	}
	
	fmt.Printf("   ✓ Email sent to: %s\n", message.To[0].Email)
	fmt.Println("   ✓ Email service examples completed")
}