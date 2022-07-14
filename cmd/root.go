package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "BoredCli",
	Short: "Get random idea what to do when you get bored",
	Long: `BoredCLI is a command line application written using Go and Cobra-Cli. 
	App uses https://www.boredapi.com/ API to download ideas to avoid getting bored.
	The project was created in the process of learning GO language and Cobra library`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
