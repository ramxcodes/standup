package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd is the base command.
var rootCmd = &cobra.Command{
	Use: "standup",
	Short: "A Simple terminal CMD that helps you out prepare quickly for standup.",
	Long: "A Simple terminal CMD that helps you out prepare quickly for standup.",
}


// Execute runs the root command. Called from main.go file
func Execute(){
	if err := rootCmd.Execute(); err != nil{
		os.Exit(1)
	}
}