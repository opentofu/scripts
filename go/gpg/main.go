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
	secretName = flag.String("secret", "GPG_PRIVATE_KEY", "GitHub organization secret name")
	visibility = flag.String("visibility", "all", "GitHub organization secret visibility")

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
		DefaultCipher:          packet.CipherAES256,
		DefaultCompressionAlgo: packet.CompressionZLIB,
		CompressionConfig: &packet.CompressionConfig{
			Level: packet.BestCompression,
		},
		RSABits: 4096,
	}

	es, err := openpgp.NewEntity(*gpgName, *gpgComment, *gpgEmail, config)
	if err != nil {
		fmt.Println(err)
		return
	}

	var keyLifetime uint32 = 3 * 365 * 24 * 60 * 60

	for _, id := range es.Identities {
		id.SelfSignature = &packet.Signature{
			CreationTime:              time.Now(),
			KeyLifetimeSecs:           &keyLifetime,
			SigType:                   packet.SigTypePositiveCert,
			PubKeyAlgo:                packet.PubKeyAlgoRSA,
			Hash:                      config.Hash(),
			IssuerKeyId:               &es.PrimaryKey.KeyId,
			PreferredHash:             []uint8{8},
			PreferredCompression:      []uint8{2},
			PreferredSymmetric:        []uint8{9},
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

	privateKeyBuffer := new(bytes.Buffer)
	err = es.SerializePrivate(privateKeyBuffer, config)
	if err != nil {
		log.Fatalf("Error serializing key: %v", err)
	}

	err = setSecret(context.Background(), privateKeyBuffer.Bytes())
	if err != nil {
		log.Fatalf("Error uploading GPG key to GitHub: %v", err)
	}
}

func setSecret(ctx context.Context, privKey []byte) error {
	log.Println("Setting secret in", *org, "...")

	g := github.NewTokenClient(nil, pat)
	k, _, err := g.Actions.GetOrgPublicKey(ctx, *org)
	if err != nil {
		return fmt.Errorf("could not get key: %w", err)
	}

	en, err := encodeWithPublicKey(base64.StdEncoding.EncodeToString(privKey), *k.Key)
	if err != nil {
		return fmt.Errorf("could not encode: %w", err)
	}

	_, err = g.Actions.CreateOrUpdateOrgSecret(ctx, *org, &github.EncryptedSecret{
		Name:           *secretName,
		KeyID:          *k.KeyID,
		EncryptedValue: en,
		Visibility:     *visibility,
	})

	return err
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
