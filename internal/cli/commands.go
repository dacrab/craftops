// Package cli provides the command-line interface for craftops
package cli

// init registers all commands and their flags
func init() {
	rootCmd.AddCommand(serverCmd, updateModsCmd, backupCmd, healthCheckCmd, initCmd)
	serverCmd.AddCommand(serverStartCmd, serverStopCmd, serverRestartCmd, serverStatusCmd)
	backupCmd.AddCommand(backupCreateCmd, backupListCmd)

	updateModsCmd.Flags().BoolVar(&forceUpdate, "force", false, "force update")
	updateModsCmd.Flags().BoolVar(&noBackup, "no-backup", false, "skip backup")
	initCmd.Flags().StringVarP(&outputPath, "output", "o", "", "config path")
	initCmd.Flags().BoolVar(&force, "force", false, "overwrite")
}
