package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/epheo/anytype-go"
	_ "github.com/epheo/anytype-go/client"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := authenticate(ctx)
	if client == nil {
		return
	}

	spaceID, err := getSpace(ctx, client)
	if err != nil {
		log.Printf("Failed to get space ID: %v", err)
		return
	}

	typeKey := createCustomType(ctx, client, spaceID)
	if typeKey == "" {
		return
	}

	objectID := createObjectOfType(ctx, client, spaceID, typeKey)
	if objectID == "" {
		return
	}

	fmt.Println("\nSuccessfully created a custom type and object!")
}

func authenticate(ctx context.Context) anytype.Client {
	fmt.Println("=== Authentication ===")

	client := anytype.NewClient(
		anytype.WithBaseURL("http://localhost:31009"),
	)

	fmt.Println("Starting authentication flow...")
	authResponse, err := client.Auth().DisplayCode(ctx, "CreateTypeExample")
	if err != nil {
		log.Printf("Failed to initiate authentication: %v", err)
		return nil
	}
	challengeID := authResponse.ChallengeID

	fmt.Println("Please check your Anytype app and enter the displayed verification code:")
	var code string
	fmt.Scanln(&code)

	tokenResponse, err := client.Auth().GetToken(ctx, challengeID, code)
	if err != nil {
		log.Printf("Authentication failed: %v", err)
		return nil
	}

	fmt.Printf("Authentication successful!\n")

	client = anytype.NewClient(
		anytype.WithBaseURL("http://localhost:31009"),
		anytype.WithAppKey(tokenResponse.AppKey),
	)

	return client
}

func getSpace(ctx context.Context, client anytype.Client) (string, error) {
	space := os.Getenv("ANYTYPE_SPACE")
	if space == "" {
		return "", errors.New("No space ID provided, missing environment variable")
	}

	return space, nil
}

func createCustomType(ctx context.Context, client anytype.Client, spaceID string) string {
	fmt.Println("\n=== Creating Custom Type ===")

	createTypeReq := anytype.CreateTypeRequest{
		Name:   "Product",
		Layout: "basic",
		Icon: &anytype.Icon{
			Format: anytype.IconFormatEmoji,
			Emoji:  "📦",
		},
		PluralName: "Products",
		Properties: []anytype.PropertyDefinition{
			{
				Key:    "price",
				Name:   "Price",
				Format: "number",
			},
			{
				Key:    "category",
				Name:   "Category",
				Format: "text",
			},
			{
				Key:    "description",
				Name:   "Description",
				Format: "text",
			},
		},
	}

	newType, err := client.Space(spaceID).Types().Create(ctx, createTypeReq)
	if err != nil {
		log.Printf("Failed to create type: %v", err)
		return ""
	}

	fmt.Printf("Created type: %s (Key: %s)\n", newType.Type.Name, newType.Type.Key)
	fmt.Printf("Type description: %s\n", newType.Type.Description)
	fmt.Printf("Type has %d property definitions\n", len(newType.Type.PropertyDefinitions))

	for _, prop := range newType.Type.PropertyDefinitions {
		fmt.Printf("  - Property: %s (%s) - Format: %s\n", prop.Name, prop.Key, prop.Format)
	}

	return newType.Type.Key
}

func createObjectOfType(ctx context.Context, client anytype.Client, spaceID, typeKey string) string {
	fmt.Println("\n=== Creating Object of Custom Type ===")

	createReq := anytype.CreateObjectRequest{
		TypeKey: typeKey,
		Name:    "Gaming Laptop",
		Icon: &anytype.Icon{
			Format: anytype.IconFormatEmoji,
			Emoji:  "💻",
		},
		Properties: []map[string]any{
			{
				"key":    "price",
				"number": 22.3,
			},
			{
				"key":  "category",
				"text": "laptop",
			},
			{
				"key":  "description",
				"text": "This is a gaming laptop with high performance specifications.",
			},
		},
	}

	newObject, err := client.Space(spaceID).Objects().Create(ctx, createReq)
	if err != nil {
		log.Printf("Failed to create object: %v", err)
		return ""
	}

	objectID := newObject.Object.ID
	fmt.Printf("Created object: %s (ID: %s)\n", newObject.Object.Name, objectID)
	fmt.Printf("Object type: %s\n", newObject.Object.Type.Name)

	objectDetails, err := client.Space(spaceID).Object(objectID).Get(ctx)
	if err != nil {
		log.Printf("Failed to get object details: %v", err)
		return objectID
	}

	fmt.Printf("Object details:\n")
	fmt.Printf("  Name: %s\n", objectDetails.Object.Name)
	fmt.Printf("  Type: %s\n", objectDetails.Object.Type.Name)
	if objectDetails.Object.Properties != nil {
		fmt.Printf("  Properties:\n")
		for key, value := range objectDetails.Object.Properties {
			fmt.Printf("    %s: %v\n", key, value)
		}
	}

	return objectID
}
