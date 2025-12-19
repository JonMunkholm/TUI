package main

import (
	"fmt"
	"os"

	"github.com/JonMunkholm/TUI/internal/application"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
)


func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found, continuing...")
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered:", r)
		}
	}()

	m, err := application.InitialModel()
	defer m.Close() // Always call Close, even on error (handles partial init)

	if err != nil {
		fmt.Printf("Failed to initialize application: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v\n", err)
		os.Exit(1)
	}
}
