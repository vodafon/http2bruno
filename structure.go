package main

import (
	"fmt"
	"os"
	"strings"
)

func DoStructure(collection, folder string) error {
	err := DoCollection(collection)
	if err != nil {
		return err
	}

	folder = strings.TrimSpace(folder)
	if folder == "" {
		return nil
	}
	return DoFolder(folder, collection)
}

func DoCollection(collection string) error {
	collection = strings.TrimSpace(collection)
	if collection == "" {
		return fmt.Errorf("-c collection name is required")
	}

	if err := os.MkdirAll("./"+collection, 0o755); err != nil {
		return fmt.Errorf("error creating directory: %v", err)
	}
	if err := os.WriteFile("./"+collection+"/collection.bru", []byte(DefaultCollectionBru()), 0o644); err != nil {
		return fmt.Errorf("error creating collection.bru: %v", err)
	}
	if err := os.WriteFile("./"+collection+"/bruno.json", []byte(DefaultBrunoJSON(collection)), 0o644); err != nil {
		return fmt.Errorf("error creating bruno.json: %w", err)
	}

	// Env
	if err := os.MkdirAll("./"+collection+"/environments", 0o755); err != nil {
		return fmt.Errorf("error creating directory: %v", err)
	}
	if err := os.WriteFile("./"+collection+"/environments/base.bru", []byte(DefaultEnvBru(collection)), 0o644); err != nil {
		return fmt.Errorf("error creating collection.bru: %v", err)
	}

	return nil
}

func DoFolder(folder, dir string) error {
	folder = strings.TrimSpace(folder)
	if folder == "" {
		return fmt.Errorf("-f folder name is required")
	}

	if err := os.MkdirAll(dir+"/"+folder, 0o755); err != nil {
		return fmt.Errorf("error creating directory: %v", err)
	}
	if err := os.WriteFile(dir+"/"+folder+"/folder.bru", []byte(DefaultFolderBru(folder)), 0o644); err != nil {
		return fmt.Errorf("error creating collection.bru: %v", err)
	}

	return nil
}
