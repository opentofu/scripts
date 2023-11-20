package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/google/go-github/v54/github"
	"golang.org/x/crypto/nacl/box"
)

var (
	owner      = flag.String("owner", "", "Owner of the repo to set the secret in")
	repoPrefix = flag.String("repo-prefix", "", "GitHub repo name prefix")

	secretKey = flag.String("secret-key", "", "Value for the secret key")

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

	flag.Parse()

	if repoPrefix == nil || *repoPrefix == "" {
		panic("repo-prefix flag not set")
	}

	c = github.NewTokenClient(context.Background(), pat)

	repos, err := getRepositories()
	if err != nil {
		log.Fatalf("failed to get repositories: %v", err)
	}

	for i := range repos {
		if err := setSecret(context.Background(), &repos[i]); err != nil {
			log.Fatalf("failed to setep secret: %v", err)
		}
	}
}

func getRepositories() ([]github.Repository, error) {
	found := make([]github.Repository, 0)
	nextPage := 1
	lastCount := 0
	for nextPage != 0 {
		log.Println("Checking page", nextPage, "...")
		repos, resp, err := c.Repositories.List(context.Background(), *owner, &github.RepositoryListOptions{
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

			if strings.HasPrefix(*repo.Name, *repoPrefix) {
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

func setSecret(ctx context.Context, repo *github.Repository) error {
	k, _, err := c.Actions.GetRepoPublicKey(ctx, repo.GetOwner().GetLogin(), repo.GetName())
	if err != nil {
		return err
	}

	en, err := encodeWithPublicKey(secretValue, *k.Key)
	if err != nil {
		return err
	}

	_, err = c.Actions.CreateOrUpdateRepoSecret(ctx, repo.GetOwner().GetLogin(), repo.GetName(), &github.EncryptedSecret{
		Name:           *secretKey,
		KeyID:          *k.KeyID,
		EncryptedValue: en,
	})
	if err != nil {
		return err
	}
	return nil
}

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
