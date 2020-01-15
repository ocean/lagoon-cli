package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var updateCmd = &cobra.Command{
	Use:     "update",
	Aliases: []string{"u"},
	Short:   "Update project, environment, or notification",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		validateToken(viper.GetString("current")) // get a new token if the current one is invalid
	},
}

func init() {
	updateCmd.AddCommand(updateSlackNotificationCmd)
	updateCmd.AddCommand(updateRocketChatNotificationCmd)
	updateCmd.AddCommand(updateUserCmd)
}
