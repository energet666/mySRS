package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Load config
	cfg, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Create temp dir for downloaded dat files
	tmpDir, err := os.MkdirTemp("", "srs-generator-*")
	if err != nil {
		log.Fatalf("Error creating temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Download latest dat files
	fmt.Println("=== Downloading latest dat files ===")
	geositePath, geoipPath, err := DownloadLatestAssets(tmpDir)
	if err != nil {
		log.Fatalf("Error downloading assets: %v", err)
	}

	// Create output directory
	if err := os.MkdirAll(cfg.OutputDir, os.ModePerm); err != nil {
		log.Fatalf("Error creating output directory: %v", err)
	}

	// Convert geoip categories
	if len(cfg.Geoip) > 0 {
		fmt.Printf("\n=== Converting geoip categories (%d) ===\n", len(cfg.Geoip))
		if err := ConvertGeoIP(geoipPath, cfg.Geoip, cfg.OutputDir); err != nil {
			log.Fatalf("Error converting geoip: %v", err)
		}
	}

	// Convert geosite categories
	if len(cfg.Geosite) > 0 {
		fmt.Printf("\n=== Converting geosite categories (%d) ===\n", len(cfg.Geosite))
		if err := ConvertGeoSite(geositePath, cfg.Geosite, cfg.OutputDir); err != nil {
			log.Fatalf("Error converting geosite: %v", err)
		}
	}

	fmt.Printf("\n=== Done! SRS files saved to %s ===\n", cfg.OutputDir)
}
