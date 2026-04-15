package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/egesut/tools-bank-go/pkg/db"
	"github.com/egesut/tools-bank-go/pkg/mcp"
)

func main() {
	log.Println("Starting MCP server...")
	
	// Initialize database
	dbPath := getDBPath()
	database, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()
	
	server := mcp.NewServer()
	server.SetDB(database)
	
	if err := server.Serve(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func getDBPath() string {
	if home := os.Getenv("HOME"); home != "" {
		return filepath.Join(home, ".tools-bank", "data.db")
	}
	return "./data.db"
}
