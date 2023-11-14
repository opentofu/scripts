package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/go-github/v54/github"
	"golang.org/x/crypto/nacl/box"
)

var (
	org  = flag.String("org", "", "GitHub organization name")
	repo = flag.String("repo-prefix", "", "GitHub repo name prefix")

	secretName = flag.String("secret-name", "GPG_PRIVATE_KEY", "GitHub organization secret name")

	envName = flag.String("env-name", "gpg", "GitHub env name")

	pat         string
	secretValue string

	c *github.Client
)

func main() {
	pat = os.Getenv("GITHUB_PAT")
	if pat == "" {
		panic("GITHUB_PAT environment variable not set")
	}

	secretValue = os.Getenv("SECRET_VALUE")
	if secretValue == "" {
		panic("SECRET_VALUE environment variable not set")
	}

	if repo == nil || *repo == "" {
		panic("repo-prefix flag not set")
	}

	flag.Parse()

	c = github.NewTokenClient(context.Background(), pat)

	repos, err := getRepositories(*org, *repo, true)
	if err != nil {
		log.Fatalf("failed to resign releases: %v", err)
	}

	for i := range repos {
		if err := setupEnv(context.Background(), &repos[i]); err != nil {
			log.Fatalf("failed to resign releases: %v", err)
		}
	}
	fmt.Println("Done!")
}

func setupEnv(ctx context.Context, repo *github.Repository) error {
	t := true
	f := false

	fmt.Println("Creating environment", *envName, "for", *repo.Name, "...")
	_, _, err := c.Repositories.CreateUpdateEnvironment(ctx, repo.GetOwner().GetLogin(), repo.GetName(), *envName, &github.CreateUpdateEnvironment{
		DeploymentBranchPolicy: &github.BranchPolicy{
			CustomBranchPolicies: &t,
			ProtectedBranches:    &f,
		},
	})
	if err != nil {
		return err
	}

	k, _, err := c.Actions.GetEnvPublicKey(ctx, int(*repo.ID), *envName)
	if err != nil {
		return fmt.Errorf("could not get key: %w", err)
	}

	en, err := encodeWithPublicKey(secretValue, *k.Key)
	if err != nil {
		return fmt.Errorf("could not encode: %w", err)
	}

	fmt.Println("Setting secret", *secretName, "for", *repo.Name, "...")
	_, err = c.Actions.CreateOrUpdateEnvSecret(ctx, int(*repo.ID), *envName, &github.EncryptedSecret{
		Name:           *secretName,
		KeyID:          *k.KeyID,
		EncryptedValue: en,
	})

	return err

}

func getRepositories(owner, nameContains string, includeForks bool) ([]github.Repository, error) {
	found := make([]github.Repository, 0)
	nextPage := 1
	lastCount := 0
	for nextPage != 0 {
		log.Println("Checking page", nextPage, "...")
		repos, resp, err := c.Repositories.List(context.Background(), owner, &github.RepositoryListOptions{
			ListOptions: github.ListOptions{
				Page:    nextPage,
				PerPage: 100,
			},
			Sort:       "full_name",
			Visibility: "public",
			Direction:  "desc",
		})
		if err != nil {
			return nil, err
		}

		for _, repo := range repos {
			if repo.Name == nil {
				continue
			}
			if repo.Fork != nil && *repo.Fork && !includeForks {
				continue
			}

			if strings.HasPrefix(*repo.Name, nameContains) {
				found = append(found, *repo)
			}
		}
		log.Println("Found", len(found), "so far...")
		nextPage = resp.NextPage

		// We sorted by name, so if we have the same count twice, we're done.
		if lastCount != 0 && len(found) == lastCount {
			break
		}

		lastCount = len(found)
	}

	return found, nil
}

// encodeWithPublicKey encrypts the given text with the given public key.
// This is required because GitHub only allows us to store secrets encrypted
// with a public key.
func encodeWithPublicKey(text string, publicKey string) (string, error) {
	// Decode the public key from base64
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return "", err
	}

	// Decode the public key
	var publicKeyDecoded [32]byte
	copy(publicKeyDecoded[:], publicKeyBytes)

	// Encrypt the secret value
	encrypted, err := box.SealAnonymous(nil, []byte(text), (*[32]byte)(publicKeyBytes), rand.Reader)

	if err != nil {
		return "", err
	}
	// Encode the encrypted value in base64
	encryptedBase64 := base64.StdEncoding.EncodeToString(encrypted)

	return encryptedBase64, nil
}
