package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ramxcodes/g-standup/internal/config"
	"github.com/spf13/cobra"
)

var days int
var author string
var setAPIKey string
var setModelName string
var enableAI bool
var disableAI bool

var standupCmd = &cobra.Command{
	Use:   "standup",
	Short: "Generate a standup report from git commits",
	Long:  `Generate a standup-ready report from recent git commits. Supports filtering by days and author.`,

	Run: func(cmd *cobra.Command, args []string) {

		cfg, err := config.Load()
		if err != nil {
			fmt.Println("Error loading config:", err)
			return
		}

		// Handle config update flags
		if setAPIKey != "" {
			cfg.APIKey = setAPIKey
			config.Save(cfg)
			fmt.Println("API key saved.")
			return
		}

		if setModelName != "" {
			cfg.ModelName = setModelName
			config.Save(cfg)
			fmt.Println("Model name updated.")
			return
		}

		if enableAI {
			cfg.AiEnabled = true
			config.Save(cfg)
			fmt.Println("AI enabled.")
			return
		}

		if disableAI {
			cfg.AiEnabled = false
			config.Save(cfg)
			fmt.Println("AI disabled.")
			return
		}

		// Build since duration
		since := fmt.Sprintf("%d days ago", days)

		// Check if current directory is a git repo
		if isGitRepo(".") {
			runGitLog(".", since)
			return
		}

		// if not a repo -> scan direct subdirectories

		entries, err := os.ReadDir(".")
		if err != nil {
			fmt.Println("Error reading directories:", err)
			return
		}

		foundRepo := false

		for _, entry := range entries {
			if entry.IsDir() {
				path := entry.Name()

				if isGitRepo(path) {
					foundRepo = true
					fmt.Printf("\n=== Repo: %s ===\n", path)
					runGitLog(path, since)
				}
			}
		}

		if !foundRepo {
			fmt.Println("(Oopsie Daisy) No git repos found in current directory!")
		}
	},
}

func init() {
	rootCmd.AddCommand(standupCmd)

	// Define flags

	// -d or --days flag (Default is last 1 day)
	standupCmd.Flags().IntVarP(
		&days,
		"days",
		"d",
		1,
		"Number of days to look back",
	)

	// -a or --author flag
	standupCmd.Flags().StringVarP(
		&author,
		"author",
		"a",
		"",
		"Filter commits by author",
	)

	standupCmd.Flags().StringVar(
		&setAPIKey,
		"set-api-key",
		"",
		"Set Gemini API key",
	)

	standupCmd.Flags().StringVar(
		&setModelName,
		"set-model-name",
		"",
		"Set Gemini model name",
	)

	standupCmd.Flags().BoolVar(
		&enableAI,
		"enable-ai",
		false,
		"Enable AI summary",
	)
	standupCmd.Flags().BoolVar(
		&disableAI,
		"disable-ai",
		false,
		"Disable AI summary",
	)
}

func isGitRepo(path string) bool {
	cmd := exec.Command("git", "-C", path, "rev-parse", "--is-inside-work-tree")

	err := cmd.Run()
	return err == nil
}

func runGitLog(path string, since string) {

	args := []string{
		"-C", path,
		"log",
		"--since=" + since,
		"--pretty=format:%h | %s | %ad | %an",
		"--date=relative",
	}

	if author != "" {
		args = append(args, "--author="+author)
	}

	cmd := exec.Command("git", args...)

	// Buffer stores stdout in memory.
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		fmt.Println("No Commits found.")
		return
	}

	output := strings.TrimSpace(out.String())

	if output == "" {
		fmt.Println("  No commits found.")
		return
	}

	// Split into actual lines
	lines := strings.Split(output, "\n")

	for _, line := range lines {

		parts := strings.Split(line, " | ")

		if len(parts) != 4 {
			continue
		}

		hash := parts[0]
		message := parts[1]
		date := parts[2]
		authorName := parts[3]

		// Trim commit message

		if len(message) > 70 {
			message = message[:67] + "..."
		}

		hashWidth := 8
		msgWidth := 70
		dateWidth := 15
		authorWidth := 15

		if len(message) > msgWidth {
			message = message[:msgWidth-3] + "..."
		}

		formatted := fmt.Sprintf(
			"  \033[36m%-*s\033[0m  %-*s  \033[33m%-*s\033[0m  \033[32m%-*s\033[0m",
			hashWidth, hash,
			msgWidth, message,
			dateWidth, date,
			authorWidth, authorName,
		)

		fmt.Println(formatted)
	}

}
