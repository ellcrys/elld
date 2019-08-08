package cmd

import (
	"fmt"
	"os"
	path "path/filepath"

	"github.com/fatih/color"

	"github.com/ellcrys/elld/accountmgr"
	"github.com/ellcrys/elld/config"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init [flags]",
	Short: "Initialize creates a default account and environment.",
	Long: `Initialize creates a default account and environment. If an environment
has already been initialized, nothing will be done.`,
	Run: func(cmd *cobra.Command, args []string) {
		setTerminateFunc(func() {
			os.Exit(-1)
		})

		pwd, _ := cmd.Flags().GetString("pwd")
		seed, _ := cmd.Flags().GetInt64("seed")
		am := accountmgr.New(path.Join(cfg.DataDir(), config.AccountDirName))
		defaultExist := accountmgr.HasDefaultAccount(am)
		if defaultExist {
			fmt.Println(color.YellowString("Client environment has already been initialized."))
			return
		}

		fmt.Println(color.HiCyanString("Default Environment Initialization..."))

		// Create the default account
		key, err := am.CreateCmd(true, seed, pwd)
		if err != nil {
			return
		}

		fmt.Println(color.GreenString("Environment has been initialized."))
		fmt.Println("Your new default account has been created, encrypted and stored.")
		fmt.Println("This account will be the default account used to interact within the network.")
		fmt.Println("Address:", color.CyanString(key.Addr().String()))
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().String("pwd", "", "Providing a password or path to a file containing a password (No interactive mode)")
	initCmd.Flags().Int64P("seed", "s", 0, "Provide a strong seed (not recommended)")
}
