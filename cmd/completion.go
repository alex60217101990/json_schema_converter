/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	utils "github.com/alex60217101990/json_schema_generator/internal/utils"
	"os"

	"github.com/spf13/cobra"
)

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	Long: fmt.Sprintf(`To load completions:

Bash:

 $ source <(%[1]s completion bash)

 # To load completions for each session, execute once:
 # Linux:
 $ %[1]s completion bash > /etc/bash_completion.d/%[1]s
 # macOS:
 $ %[1]s completion bash > $(brew --prefix)/etc/bash_completion.d/%[1]s

Zsh:

 # If shell completion is not already enabled in your environment,
 # you will need to enable it.  You can execute the following once:

 $ echo "autoload -U compinit; compinit" >> ~/.zshrc

 # To load completions for each session, execute once:
 $ %[1]s completion zsh > "${fpath[1]}/_%[1]s"

 # You will need to start a new shell for this setup to take effect.

fish:

 $ %[1]s completion fish | source

 # To load completions for each session, execute once:
 $ %[1]s completion fish > ~/.config/fish/completions/%[1]s.fish

PowerShell:

 PS> %[1]s completion powershell | Out-String | Invoke-Expression

 # To load completions for every new session, run:
 PS> %[1]s completion powershell > %[1]s.ps1
 # and source this file from your PowerShell profile.
`, rootCmd.Root().Name()),
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactValidArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		log = utils.InitLogger(true)
		var err error
		switch args[0] {
		case "bash":
			err = cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			err = cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			err = cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			err = cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}

		if err != nil {
			log.Error().Stack().Err(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// completionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// completionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
