package cmd

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)


var days int
var author string

var standupCmd = &cobra.Command{
	Use: "standup",
	Short: "Generate a standup report from git commits",
	Long: `Generate a standup-ready report from recent git commits. Supports filtering by days and author.`,

	Run: func(cmd *cobra.Command, args []string){
		fmt.Println("Running Standup command...")
		
		// If author is empty, default to current git user
		if author == "" {
			author = getGitUser()
		}

		// Build since duration
		since := fmt.Sprintf("%d days ago", days)

		// Preprare git command

		gitCmd := exec.Command(
			"git",
			"log",
			"--since="+since,
			"--author"+author,
			"--pretty=format:%h | %ad | %an | %s",
			"--date=relative",
		)


		// Buffer to capture output

		var out bytes.Buffer
		gitCmd.Stdout = &out

		// Execute command

		err := gitCmd.Run()
		if err != nil{
			fmt.Println("Error running git log:", err)
			return
		}

		// Print raw git output

		fmt.Println("\nCommits:\n")
		fmt.Println(out.String())
	},
}

func init(){
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
}


func getGitUser() string {
	cmd := exec.Command("git", "config", "user.name")

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return ""
	}

	return string(bytes.TrimSpace(out.Bytes()))
}