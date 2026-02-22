package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const version = "1"

const helpASCII = `
                            ╱|、
                          (˚ˎ 。7  
                           |、˜〵          
                          じしˍ,)ノ
		made with ♡ by Ram
`

const versionASCII = `
⠀⠀⠀⠀⠀⠀⠀⢠⢰⣦⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣤⣿⡏⣦⡀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⣐⣯⣟⡿⣿⣧⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣼⣿⡋⠀⣿⣿⠇⠀⠀⠀⠀
⠀⠀⠀⠀⠀⡐⣮⣾⡝⠈⠻⡯⣿⣢⢄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⢴⣿⣿⠃⠀⠀⢹⣿⡖⠄⠀⠀⠀
⠀⠀⠀⠀⠀⡮⡿⠇⠀⠀⠀⠈⠺⣿⣯⢦⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣐⣵⣿⡗⠁⠀⠀⠀⠈⣼⢿⡔⠀⠀⠀
⠀⠀⠀⢠⣯⣽⠃⠀⠀⠀⠀⠀⠀⠀⠙⢿⣿⡄⠀⠀⠒⡾⣽⣤⣦⣦⣦⣶⣶⣶⣦⡍⠁⠀⠜⣻⣿⠋⠀⠀⠀⠀⠀⠀⠈⣿⣿⡁⠀⠀
⠀⠀⢀⢯⣹⠃⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠉⠁⠀⠀⠀⠉⠛⠛⠋⠋⠉⠉⠉⠛⠋⠉⠁⠀⠀⠡⠃⠀⠀⠀⠀⠀⠀⠀⠀⢐⡾⣯⠀⠀
⠀⠀⣎⢳⠟⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣀⣀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣠⣄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⢿⢟⠠⠀
⠀⢰⡿⣽⠃⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢏⣿⣿⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣾⣿⡿⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠸⣏⡐
⠀⡿⠩⡇⠀⠀⢀⣀⣀⣄⡀⠀⠀⠀⠀⢸⢸⣿⠅⠀⠀⠀⠀⠀⢀⣀⣀⡀⠀⠀⠀⠀⠀⣿⣿⣷⠃⠀⠀⡀⣦⣤⣴⢦⠀⠀⠈⣯⡦⠂
⢸⡏⣶⠁⠀⠀⣸⣿⣿⣿⣿⡳⠀⠀⠀⠸⣿⠋⠀⠀⠀⠀⠀⠀⠘⣟⣿⠟⠁⠀⠀⠀⠀⢻⡿⡟⠀⠀⢺⣽⣿⣮⣿⣿⠇⠀⠀⢿⣟
⣿⡇⡯⠀⠀⠀⠀⠈⠉⠉⠈⠁⠀⠀⠀⠀⠈⠀⠀⠀⠀⠀⠀⢀⣾⠇⠁⠙⢿⡦⠀⠀⠀⠀⠈⠁⠀⠀⠀⠀⠉⠉⠉⠁⠀⠀⡇⠘⣓

Version 1.0.0
`

const helpFooter = `
   Having issues? Connect with Ram — ramx.in
`

// rootCmd is the base command.
var rootCmd = &cobra.Command{
	Use:   "standup",
	Short: "A Simple terminal CMD that helps you out prepare quickly for standup.",
	Long:  "A Simple terminal CMD that helps you out prepare quickly for standup.",
	Run: func(c *cobra.Command, args []string) {
		if ok, _ := c.Flags().GetBool("version"); ok {
			fmt.Printf(versionASCII, version)
			return
		}
		RunStandup(c, args)
	},
}

func init() {
	rootCmd.Flags().BoolP("version", "v", false, "show version")
	rootCmd.SetHelpTemplate(helpASCII + `Usage:{{if .Runnable}}
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
}

// Execute runs the root command. Called from main.go file
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}