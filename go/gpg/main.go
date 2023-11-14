package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
	"github.com/google/go-github/v54/github"
	"golang.org/x/crypto/nacl/box"
)

var (
	org        = flag.String("org", "", "GitHub organization name")
	repo       = flag.String("repo", "scripts", "GitHub repo name")
	secretName = flag.String("secret", "GPG_PRIVATE_KEY", "GitHub organization secret name")

	gpgName    = flag.String("gpg-name", "opentofu", "GPG name")
	gpgComment = flag.String("gpg-comment", "This is the key used to sign opentofu providers", "GPG comment")
	gpgEmail   = flag.String("gpg-email", "your.email@example.com", "GPG comment")

	pat string
)

func main() {
	pat = os.Getenv("GITHUB_PAT")
	if pat == "" {
		panic("GITHUB_PAT environment variable not set")
	}

	flag.Parse()

	config := &packet.Config{
		DefaultCipher:          packet.CipherAES256,    // [https://datatracker.ietf.org/doc/html/rfc4880#section-9.2]
		DefaultCompressionAlgo: packet.CompressionZLIB, // [https://datatracker.ietf.org/doc/html/rfc4880#section-9.3]
		CompressionConfig: &packet.CompressionConfig{
			Level: packet.BestCompression, // [https://datatracker.ietf.org/doc/html/rfc4880#section-9.3]
		},
		RSABits: 4096, // 4096-bit RSA key [https://datatracker.ietf.org/doc/html/rfc4880#section-
	}

	// Generate a new Entity. Will contain a new key pair and a self-signed identity
	// which we'll adjust later.
	es, err := openpgp.NewEntity(*gpgName, *gpgComment, *gpgEmail, config)
	if err != nil {
		fmt.Println(err)
		return
	}

	var keyLifetime uint32 = 3 * 365 * 24 * 60 * 60 // 3 years

	// There should be only one identity, but let's iterate over them anyway.
	for _, id := range es.Identities {
		// SelfSignature is a key component in OpenPGP for identity verification.
		// It is a signature made by the owner of the identity on their own public key.
		//
		// After setting the SelfSignature, the user ID (name and email) is signed with the entity's
		// primary key and private key, creating a verifiable signature for future reference.
		//
		// This process is crucial in OpenPGP as it allows the owner of an identity to claim
		// ownership of a public key, and for others to verify that claim.
		id.SelfSignature = &packet.Signature{
			CreationTime:              time.Now(),
			KeyLifetimeSecs:           &keyLifetime,
			SigType:                   packet.SigTypePositiveCert, // [https://datatracker.ietf.org/doc/html/rfc4880#section-5.2.1]
			PubKeyAlgo:                packet.PubKeyAlgoRSA,       // [https://datatracker.ietf.org/doc/html/rfc4880#section-9.1]
			Hash:                      config.Hash(),
			IssuerKeyId:               &es.PrimaryKey.KeyId,                   // Key ID of the signing key
			PreferredHash:             []uint8{8},                             // 8 for SHA256 (Encrypt or Sign) [https://datatracker.ietf.org/doc/html/rfc4880#section-9.4]
			PreferredCompression:      []uint8{uint8(packet.CompressionZLIB)}, // [https://datatracker.ietf.org/doc/html/rfc4880#section-9.3]
			PreferredSymmetric:        []uint8{uint8(packet.CipherAES256)},    // [https://datatracker.ietf.org/doc/html/rfc4880#section-9.2]
			IsPrimaryId:               new(bool),
			FlagsValid:                true,
			FlagCertify:               false,
			FlagEncryptCommunications: false,
			FlagEncryptStorage:        false,
		}

		err := id.SelfSignature.SignUserId(id.UserId.Id, es.PrimaryKey, es.PrivateKey, config)
		if err != nil {
			log.Fatalf("Error signing identity: %v", err)
		}
	}

	fmt.Println("Generating GPG key...")
	privateKeyBuffer := new(bytes.Buffer)
	err = es.SerializePrivate(privateKeyBuffer, config)
	if err != nil {
		log.Fatalf("Error serializing key: %v", err)
	}

	fmt.Println("Setting secret in", *org+"/"+*repo+"...")
	err = setSecret(context.Background(), privateKeyBuffer.Bytes())
	if err != nil {
		log.Fatalf("Error uploading GPG key to GitHub: %v", err)
	}

	var pb bytes.Buffer
	if err = es.Serialize(&pb); err != nil {
		log.Fatalf("Error serializing public key: %v", err)
	}

	fmt.Println("")
	fmt.Println("PUBLIC KEY BASE64:")
	fmt.Println(base64.StdEncoding.EncodeToString(pb.Bytes()))
}

func setSecret(ctx context.Context, privKey []byte) error {
	g := github.NewTokenClient(nil, pat)
	k, _, err := g.Actions.GetRepoPublicKey(ctx, *org, *repo)
	if err != nil {
		return fmt.Errorf("could not get key: %w", err)
	}

	en, err := encodeWithPublicKey(base64.StdEncoding.EncodeToString(privKey), *k.Key)
	if err != nil {
		return fmt.Errorf("could not encode: %w", err)
	}

	_, err = g.Actions.CreateOrUpdateRepoSecret(ctx, *org, *repo, &github.EncryptedSecret{
		Name:           *secretName,
		KeyID:          *k.KeyID,
		EncryptedValue: en,
	})

	return err
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
