package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ramxcodes/homebrew-standup/internal/ai"
	"github.com/ramxcodes/homebrew-standup/internal/config"
	"github.com/spf13/cobra"
)

var days int
var author string
var setAPIKey string
var setModelName string
var showAPIKey bool
var removeAPIKey bool
var enableAI bool
var disableAI bool

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

var standupCmd = &cobra.Command{
	Use:   "standup",
	Short: "Generate a standup report from git commits",
	Long:  `Generate a standup-ready report from recent git commits. Supports filtering by days and author.`,
	Run:   func(c *cobra.Command, args []string) { RunStandup(c, args) },
}

func RunStandup(cmd *cobra.Command, args []string) {
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

		if showAPIKey {
			if cfg.APIKey == "" {
				fmt.Println("  No API key set.")
			} else {
				// Mask middle, show last 4 for verification
				k := cfg.APIKey
				if len(k) <= 8 {
					fmt.Println("  " + k)
				} else {
					fmt.Printf("  ****...%s\n", k[len(k)-4:])
				}
			}
			return
		}

		if removeAPIKey {
			cfg.APIKey = ""
			config.Save(cfg)
			fmt.Println("  API key removed.")
			return
		}

		// Default author to current git user when not set
		authFilter := author
		if authFilter == "" {
			authFilter = getDefaultAuthor(".")
		}

		// Build since duration
		since := fmt.Sprintf("%d days ago", days)

		var allRawLines []string

		// Check if current directory is a git repo
		if isGitRepo(".") {
			display, raw, has := runGitLog(".", since, authFilter)
			for _, line := range display {
				fmt.Println(line)
			}
			allRawLines = append(allRawLines, raw...)
			if cfg.AiEnabled && cfg.APIKey != "" && has {
				printAISummary(cfg, allRawLines)
			}
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
					display, raw, _ := runGitLog(path, since, authFilter)
					for _, line := range display {
						fmt.Println(line)
					}
					allRawLines = append(allRawLines, raw...)
				}
			}
		}

		if !foundRepo {
			fmt.Println("(Oopsie Daisy) No git repos found in current directory!")
			return
		}

		if cfg.AiEnabled && cfg.APIKey != "" && len(allRawLines) > 0 {
			fmt.Println()
			printAISummary(cfg, allRawLines)
		}
}

func init() {
	rootCmd.AddCommand(standupCmd)

	standupCmd.SetHelpTemplate(helpASCII + `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}
` + helpFooter + "\n")

	// -d or --days flag (Default is last 1 day)
	rootCmd.Flags().IntVarP(
		&days,
		"days",
		"d",
		1,
		"Number of days to look back",
	)

	// -a or --author flag
	rootCmd.Flags().StringVarP(
		&author,
		"author",
		"a",
		"",
		"Filter commits by author",
	)

	rootCmd.Flags().StringVar(
		&setAPIKey,
		"set-api-key",
		"",
		"Set Gemini API key",
	)

	rootCmd.Flags().StringVar(
		&setModelName,
		"set-model-name",
		"",
		"Set Gemini model name",
	)

	rootCmd.Flags().BoolVar(
		&showAPIKey,
		"show-api-key",
		false,
		"Show stored API key (masked)",
	)

	rootCmd.Flags().BoolVar(
		&removeAPIKey,
		"remove-api-key",
		false,
		"Remove stored API key",
	)

	rootCmd.Flags().BoolVar(
		&enableAI,
		"enable-ai",
		false,
		"Enable AI summary",
	)
	rootCmd.Flags().BoolVar(
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

func getDefaultAuthor(repoPath string) string {
	cmd := exec.Command("git", "-C", repoPath, "config", "user.name")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return ""
	}
	return strings.TrimSpace(out.String())
}

// runGitLog returns (display lines for terminal, raw lines for AI, hasCommits).
func runGitLog(path, since, authorFilter string) (display []string, raw []string, hasCommits bool) {
	args := []string{
		"-C", path,
		"log",
		"--since=" + since,
		"--pretty=format:%h | %s | %ad | %an",
		"--date=relative",
	}
	if authorFilter != "" {
		args = append(args, "--author="+authorFilter)
	}

	cmd := exec.Command("git", args...)
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return []string{"  No commits found."}, nil, false
	}

	output := strings.TrimSpace(out.String())
	if output == "" {
		return []string{"  No commits found."}, nil, false
	}

	lines := strings.Split(output, "\n")
	hashWidth := 8
	msgWidth := 70
	dateWidth := 15
	authorWidth := 15

	for _, line := range lines {
		parts := strings.Split(line, " | ")
		if len(parts) != 4 {
			continue
		}
		raw = append(raw, line)

		hash := parts[0]
		message := parts[1]
		date := parts[2]
		authorName := parts[3]
		if len(message) > msgWidth {
			message = message[:msgWidth-3] + "..."
		}

		formatted := fmt.Sprintf(
			"  \033[36m%-*s\033[0m  %-*s  \033[33m%-*s\033[0m  \033[32m%-*s\033[0m",
			hashWidth, hash, msgWidth, message, dateWidth, date, authorWidth, authorName,
		)
		display = append(display, formatted)
	}

	return display, raw, len(raw) > 0
}

func runWithLoader(msg string, fn func() (string, error)) (string, error) {
	done := make(chan struct{})
	var result string
	var err error
	go func() {
		result, err = fn()
		close(done)
	}()

	i := 0
	for {
		select {
		case <-done:
			fmt.Print("\r\033[K") // clear line
			return result, err
		default:
			frame := spinnerFrames[i%len(spinnerFrames)]
			fmt.Printf("\r  %s %s  ", msg, frame)
			i++
			time.Sleep(80 * time.Millisecond)
		}
	}
}

func printAISummary(cfg *config.Config, rawLines []string) {
	prompt := `Summarize these git commits into a short standup update: what was done recently and what's in progress. Keep it casual and brief (1–3 sentences).

Commits:
` + strings.Join(rawLines, "\n")

	fmt.Println("\n")
	summary, err := runWithLoader("☆ Generating summary", func() (string, error) {
		return ai.GenerateSummary(cfg.APIKey, cfg.ModelName, prompt)
	})

	if err != nil {
		fmt.Printf("  \033[31mAI summary failed: %v\033[0m\n", err)
		return
	}

	summary = strings.TrimSpace(summary)
	fmt.Println("\033[1mAI GENERATED SUMMARY -\033[0m")
	fmt.Println()
	fmt.Println("  " + summary)
	fmt.Println()
}
