package cmd

import (
	"bytes"
	"fmt"
	"os"
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
		// Build since duration
		since := fmt.Sprintf("%d days ago", days)

		// Check if current directory is a git repo
		if isGitRepo("."){
			runGitLog(".", since)
			return
		}

		// if not a repo -> scan direct subdirectories

		entries, err := os.ReadDir(".")
		if err != nil{
			fmt.Println("Error reading directories:", err)
			return
		}

		foundRepo := false

		for _, entry := range entries{
			if entry.IsDir(){
				path := entry.Name()

				if isGitRepo(path){
					foundRepo = true
					fmt.Printf("\n=== Repo: %s ===\n", path)
					runGitLog(path,since)
				}
			}
		}

		if !foundRepo{
			fmt.Println("(Oopsie Daisy) No git repos found in current directory!")
		}
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

func isGitRepo(path string) bool {
	cmd := exec.Command("git", "-C", path, "rev-parse", "--is-inside-work-tree")

	err := cmd.Run()
	return err == nil
}

func runGitLog (path string, since string){

	args := []string{
		"-C", path,
		"log",
		"--since=" + since,
		"--pretty=format:%h | %s | %ad | %an",
		"--date=relative",
	}

	if author != ""{
		args = append(args, "--author="+author)
	}

	cmd := exec.Command("git", args...)

	var out bytes.Buffer
	cmd.Stdout = &out
	

	err := cmd.Run()
	if err != nil{
		fmt.Println("No Commits found.")
		return
	}

	fmt.Println(out.String())
}