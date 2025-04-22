package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for Envi CLI.
To load completions:

Bash:
  $ source <(envi completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ envi completion bash > /etc/bash_completion.d/envi
  # macOS:
  $ envi completion bash > $(brew --prefix)/etc/bash_completion.d/envi

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ envi completion zsh > "${fpath[1]}/_envi"

Fish:
  $ envi completion fish | source

  # To load completions for each session, execute once:
  $ envi completion fish > ~/.config/fish/completions/envi.fish

PowerShell:
  PS> envi completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> envi completion powershell > envi.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactValidArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
	},
}

// InitCompletionCommand initializes the completion command
func InitCompletionCommand() {
	rootCmd.AddCommand(completionCmd)
} 