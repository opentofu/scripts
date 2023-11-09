package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v54/github"
)

var (
	owner = flag.String("owner", "", "GitHub repo owner name")
	repo  = flag.String("repo", "", "GitHub repo name")

	fingerprint = flag.String("fingerprint", "", "GPG fingerprint")
)

var c *github.Client

func main() {
	flag.Parse()

	pat := os.Getenv("GITHUB_PAT")
	if pat == "" {
		panic("GITHUB_PAT environment variable not set")
	}

	c = github.NewTokenClient(nil, pat)

	err := resignReleases(context.Background(), *fingerprint)
	if err != nil {
		log.Fatalf("failed to resign releases: %v", err)
	}
}

func resignReleases(ctx context.Context, fingerprint string) error {
	tmpdir, err := os.MkdirTemp("", "")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpdir)
	log.Printf("tmpdir: %s", tmpdir)

	releases, err := getReleases(ctx, "v")
	if err != nil {
		return fmt.Errorf("could not list %s/%s releases: %w", *owner, *repo, err)
	}
	log.Printf("releases: %d", len(releases))

	for _, release := range releases {
		log.Printf("release: %s", release.GetTagName())
		assets, _, err := c.Repositories.ListReleaseAssets(ctx, *owner, *repo, release.GetID(), &github.ListOptions{})
		if err != nil {
			return fmt.Errorf("could not list %s/%s release %s assets: %w", *owner, *repo, release.GetTagName(), err)
		}
		log.Printf("assets: %d", len(assets))
		var (
			checksumAssetID   int64 = -1
			checksumAssetName string
			signatureAssetID  int64 = -1
		)
		for _, asset := range assets {
			log.Printf("asset: %s=%d", asset.GetName(), asset.GetID())
			if strings.HasSuffix(asset.GetName(), "_SHA256SUMS") {
				checksumAssetID = asset.GetID()
				checksumAssetName = asset.GetName()
			}
			if strings.HasSuffix(asset.GetName(), "_SHA256SUMS.sig") {
				signatureAssetID = asset.GetID()
			}
		}
		log.Printf("checksum=%d,signature=%d", checksumAssetID, signatureAssetID)
		if checksumAssetID < 0 || signatureAssetID < 0 {
			return fmt.Errorf("could not find %s/%s release %s assets, checksum=%t,signature=%t", *owner, *repo, release.GetTagName(), checksumAssetID < 0, signatureAssetID < 0)
		}
		log.Printf("download asset %d as %s", checksumAssetID, checksumAssetName)
		if err := downloadAsset(ctx, checksumAssetID, filepath.Join(tmpdir, checksumAssetName)); err != nil {
			return fmt.Errorf("could not download %s/%s release %s checksum asset %d: %w", *owner, *repo, release.GetTagName(), checksumAssetID, err)
		}
		log.Printf("sign asset %s", checksumAssetName)
		signatureFilename, err := sign(fingerprint, tmpdir, checksumAssetName)
		if err != nil {
			return err
		}
		log.Printf("delete asset %d", signatureAssetID)
		if err := deleteAsset(ctx, signatureAssetID); err != nil {
			return fmt.Errorf("could not delete %s/%s release %s asset %d: %w", *owner, *repo, release.GetTagName(), signatureAssetID, err)
		}
		log.Printf("upload asset %s", signatureFilename)
		if err := uploadAsset(ctx, release.GetID(), filepath.Join(tmpdir, signatureFilename), signatureFilename); err != nil {
			return fmt.Errorf("could not upload %s/%s release %s asset %s: %w", *owner, *repo, release.GetTagName(), signatureFilename, err)
		}
	}

	return nil
}

func getReleases(ctx context.Context, prefix string) ([]*github.RepositoryRelease, error) {
	var repoReleases []*github.RepositoryRelease
	page := 1
	for {
		releases, resp, err := c.Repositories.ListReleases(ctx, *owner, *repo, &github.ListOptions{
			Page:    page,
			PerPage: 99,
		})
		if err != nil {
			return nil, err
		}
		if prefix != "" {
			for _, release := range releases {
				if strings.HasPrefix(release.GetTagName(), prefix) {
					repoReleases = append(repoReleases, release)
				}
			}
		} else {
			repoReleases = append(repoReleases, releases...)
		}
		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}
	return repoReleases, nil
}

func downloadAsset(ctx context.Context, id int64, filename string) error {
	rc, _, err := c.Repositories.DownloadReleaseAsset(ctx, *owner, *repo, id, http.DefaultClient)
	if err != nil {
		return fmt.Errorf("could not download asset: %w", err)
	}
	defer rc.Close()
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("could not create file %s: %w", filename, err)
	}
	defer f.Close()
	if _, err := io.Copy(f, rc); err != nil {
		return fmt.Errorf("could not copy asset data: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("could not close file %s: %w", filename, err)
	}
	return nil
}

func deleteAsset(ctx context.Context, id int64) error {
	_, err := c.Repositories.DeleteReleaseAsset(ctx, *owner, *repo, id)
	if err != nil {
		return fmt.Errorf("could not download asset: %w", err)
	}
	return nil
}

func uploadAsset(ctx context.Context, id int64, filename, name string) error {
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("could not open %s: %w", filename, err)
	}
	defer f.Close()
	_, _, err = c.Repositories.UploadReleaseAsset(ctx, *owner, *repo, id, &github.UploadOptions{Name: name}, f)
	if err != nil {
		return fmt.Errorf("could not upload asset: %w", err)
	}
	return nil
}

func sign(fingerprint string, dir, filename string) (string, error) {
	signatureFilename := filename + ".sig"
	cmd := exec.Command("gpg", "--batch", "--local-user", fingerprint, "--output", signatureFilename, "--detach-sign", filename)
	cmd.Dir = dir
	b, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed '%s': %w", string(b), err)
	}
	return signatureFilename, nil
}
