package main

import (
	"context"
	"log"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/iterator"
)

func main() {
	ctx := context.Background()

	// Initialize the Firebase Admin SDK.
	// It automatically uses the service account credentials from the
	// GOOGLE_APPLICATION_CREDENTIALS environment variable.
	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}

	client, err := app.Auth(ctx)
	if err != nil {
		log.Fatalf("error getting Auth client: %v\n", err)
	}

	log.Println("Successfully connected to Firebase. Starting user cleanup...")

	var uidsToDelete []string
	pager := iterator.NewPager(client.Users(ctx, ""), 1000, "")

	for {
		var users []*auth.ExportedUserRecord
		nextPageToken, err := pager.NextPage(&users)
		if err != nil {
			log.Fatalf("error listing users: %v", err)
		}

		for _, u := range users {
			if strings.HasPrefix(u.Email, "test-") {
				uidsToDelete = append(uidsToDelete, u.UID)
			}
		}

		if nextPageToken == "" {
			break
		}
	}

	if len(uidsToDelete) == 0 {
		log.Println("✅ No test users with 'test-' prefix found.")
		return
	}

	log.Printf("Found %d test users to delete.", len(uidsToDelete))

	// Delete the users in batches of up to 1000.
	// The DeleteUsers function handles this batching internally if the list is large.
	result, err := client.DeleteUsers(ctx, uidsToDelete)
	if err != nil {
		log.Fatalf("error deleting users: %v", err)
	}

	log.Printf("✅ Successfully deleted %d users.", result.SuccessCount)
	if result.FailureCount > 0 {
		log.Printf("⚠️ Failed to delete %d users.", result.FailureCount)
		for _, err := range result.Errors {
			log.Printf(" - error: %v", err.Reason)
		}
	}
}
