package gcp

import (
	"context"
	"fmt"
	"log"
	"os"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

// UpdateSecret updates a secret in GCP Secret Manager
func UpdateSecret(secretName, secretValue string) error {
	projectID := os.Getenv("GCP_PROJECT")
	if projectID == "" {
		return fmt.Errorf("GCP_PROJECT environment variable not set")
	}

	autoUpdate := os.Getenv("AUTO_UPDATE_SECRETS")
	if autoUpdate != "true" {
		log.Printf("[SECRETS] AUTO_UPDATE_SECRETS not enabled, skipping update for %s", secretName)
		return nil
	}

	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create Secret Manager client: %w", err)
	}
	defer client.Close()

	// Add the new secret version
	req := &secretmanagerpb.AddSecretVersionRequest{
		Parent: fmt.Sprintf("projects/%s/secrets/%s", projectID, secretName),
		Payload: &secretmanagerpb.SecretPayload{
			Data: []byte(secretValue),
		},
	}

	version, err := client.AddSecretVersion(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to add secret version: %w", err)
	}

	log.Printf("[SECRETS] Successfully updated secret %s to version %s", secretName, version.Name)
	return nil
}

// UpdateYotoTokens updates both access and refresh tokens in Secret Manager
func UpdateYotoTokens(accessToken, refreshToken string) error {
	var errs []error

	if accessToken != "" {
		if err := UpdateSecret("yoto-access-token", accessToken); err != nil {
			log.Printf("[SECRETS] Failed to update yoto-access-token: %v", err)
			errs = append(errs, err)
		}
	}

	if refreshToken != "" {
		if err := UpdateSecret("yoto-refresh-token", refreshToken); err != nil {
			log.Printf("[SECRETS] Failed to update yoto-refresh-token: %v", err)
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to update %d secret(s)", len(errs))
	}

	return nil
}
