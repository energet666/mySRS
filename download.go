package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const (
	releasesAPIURL = "https://api.github.com/repos/runetfreedom/russia-v2ray-rules-dat/releases/latest"
)

type releaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type releaseResponse struct {
	TagName string         `json:"tag_name"`
	Assets  []releaseAsset `json:"assets"`
}

// DownloadLatestAssets downloads the latest geosite.dat and geoip.dat from GitHub releases.
// Returns paths to the downloaded files in tmpDir.
func DownloadLatestAssets(tmpDir string) (geositePath, geoipPath string, err error) {
	// Fetch latest release info
	resp, err := http.Get(releasesAPIURL)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release releaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", "", fmt.Errorf("failed to parse release info: %w", err)
	}

	fmt.Printf("Latest release: %s\n", release.TagName)

	// Find download URLs for geosite.dat and geoip.dat
	var geositeURL, geoipURL string
	for _, asset := range release.Assets {
		switch asset.Name {
		case "geosite.dat":
			geositeURL = asset.BrowserDownloadURL
		case "geoip.dat":
			geoipURL = asset.BrowserDownloadURL
		}
	}

	if geositeURL == "" {
		return "", "", fmt.Errorf("geosite.dat not found in release assets")
	}
	if geoipURL == "" {
		return "", "", fmt.Errorf("geoip.dat not found in release assets")
	}

	// Download both files
	geositePath = filepath.Join(tmpDir, "geosite.dat")
	if err := downloadFile(geositeURL, geositePath); err != nil {
		return "", "", fmt.Errorf("failed to download geosite.dat: %w", err)
	}
	fmt.Printf("Downloaded geosite.dat (%s)\n", geositeURL)

	geoipPath = filepath.Join(tmpDir, "geoip.dat")
	if err := downloadFile(geoipURL, geoipPath); err != nil {
		return "", "", fmt.Errorf("failed to download geoip.dat: %w", err)
	}
	fmt.Printf("Downloaded geoip.dat (%s)\n", geoipURL)

	return geositePath, geoipPath, nil
}

func downloadFile(url, destPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d for %s", resp.StatusCode, url)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
