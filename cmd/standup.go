package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ramxcodes/homebrew-standup/internal/ai"
	"github.com/ramxcodes/homebrew-standup/internal/config"
	"github.com/spf13/cobra"
)

const (
	colorReset   = "\033[0m"
	colorDim     = "\033[2m"
	colorBold    = "\033[1m"
	colorCyan    = "\033[36m"
	colorYellow  = "\033[33m"
	colorGreen   = "\033[32m"
	colorMagenta = "\033[35m"
	colorGray    = "\033[90m"
	colorRed     = "\033[31m"
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
			fmt.Printf("%sError loading config: %v%s\n", colorRed, err, colorReset)
			return
		}

		// Handle config update flags
		if setAPIKey != "" {
			cfg.APIKey = setAPIKey
			config.Save(cfg)
			fmt.Printf("%sAPI key saved.%s\n", colorGreen, colorReset)
			return
		}

		if setModelName != "" {
			cfg.ModelName = setModelName
			config.Save(cfg)
			fmt.Printf("%sModel name updated.%s\n", colorGreen, colorReset)
			return
		}

		if enableAI {
			cfg.AiEnabled = true
			config.Save(cfg)
			fmt.Printf("%sAI enabled.%s\n", colorGreen, colorReset)
			return
		}

		if disableAI {
			cfg.AiEnabled = false
			config.Save(cfg)
			fmt.Printf("%sAI disabled.%s\n", colorGreen, colorReset)
			return
		}

		if showAPIKey {
			if cfg.APIKey == "" {
				fmt.Printf("  %sNo API key set.%s\n", colorDim, colorReset)
			} else {
				k := cfg.APIKey
				if len(k) <= 8 {
					fmt.Printf("  %s%s%s\n", colorCyan, k, colorReset)
				} else {
					fmt.Printf("  %s****...%s%s\n", colorDim, k[len(k)-4:], colorReset)
				}
			}
			return
		}

		if removeAPIKey {
			cfg.APIKey = ""
			config.Save(cfg)
			fmt.Printf("  %sAPI key removed.%s\n", colorGreen, colorReset)
			return
		}

		if cfg.AiEnabled && cfg.APIKey == "" {
			fmt.Printf("%sAI is enabled but no API key is set.%s\n", colorYellow, colorReset)
			fmt.Printf("%sSet one with %s--set-api-key <key>%s or disable AI with %s--disable-ai%s\n", colorDim, colorCyan, colorDim, colorCyan, colorReset)
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
			absPath, _ := filepath.Abs(".")
			display, raw, has := runGitLog(".", since, authFilter)
			printRepoOutput(absPath, display, authFilter, has)
			allRawLines = append(allRawLines, raw...)
			if cfg.AiEnabled && cfg.APIKey != "" && has {
				printAISummary(cfg, allRawLines)
			}
			return
		}

		// if not a repo -> scan direct subdirectories
		entries, err := os.ReadDir(".")
		if err != nil {
			fmt.Printf("%sError reading directories: %v%s\n", colorGray, err, colorReset)
			return
		}

		type repoResult struct {
			absPath   string
			display   []string
			raw       []string
			hasCommits bool
		}
		var noCommitRepos []string
		var commitRepos []repoResult

		for _, entry := range entries {
			if entry.IsDir() {
				path := entry.Name()
				if isGitRepo(path) {
					absPath, _ := filepath.Abs(path)
					display, raw, has := runGitLog(path, since, authFilter)
					if !has {
						noCommitRepos = append(noCommitRepos, absPath)
					} else {
						commitRepos = append(commitRepos, repoResult{absPath, display, raw, has})
						allRawLines = append(allRawLines, raw...)
					}
				}
			}
		}

		if len(noCommitRepos) == 0 && len(commitRepos) == 0 {
			fmt.Printf("%s(Oopsie Daisy) No git repos found in current directory!%s\n", colorYellow, colorReset)
			return
		}

		// Show no-commit repos first, then repos with commits
		for _, absPath := range noCommitRepos {
			printRepoOutput(absPath, nil, authFilter, false)
		}
		for _, r := range commitRepos {
			printRepoOutput(r.absPath, r.display, authFilter, true)
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

func printRepoOutput(absPath string, display []string, authorName string, hasCommits bool) {
	fmt.Printf("%s%s%s\n", colorMagenta, absPath, colorReset)
	if !hasCommits {
		who := authorName
		if who == "" {
			who = "you"
		}
		fmt.Printf("%sNo commits from %s during this period.%s\n", colorDim, who, colorReset)
		return
	}
	for _, line := range display {
		fmt.Println(line)
	}
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
// When no commits: display is nil. Format: "hash - message (date) <author>" with colors.
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
		return nil, nil, false
	}

	output := strings.TrimSpace(out.String())
	if output == "" {
		return nil, nil, false
	}

	lines := strings.Split(output, "\n")
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

		formatted := fmt.Sprintf("%s%s%s - %s (%s%s%s) %s<%s>%s",
			colorCyan, hash, colorReset,
			message,
			colorYellow, date, colorReset,
			colorGreen, authorName, colorReset)
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
			fmt.Printf("\r  %s%s %s%s  ", colorCyan, msg, frame, colorReset)
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
		fmt.Printf("  %sAI summary failed: %v%s\n", colorRed, err, colorReset)
		return
	}

	summary = strings.TrimSpace(summary)
	fmt.Printf("\n%s%sAI GENERATED SUMMARY%s%s\n\n", colorBold, colorCyan, colorReset, colorBold)
	fmt.Println("  " + summary)
	fmt.Println()
}
