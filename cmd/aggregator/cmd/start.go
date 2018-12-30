// Copyright © 2018 Emando B.V.

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// startCmd represents the start command.
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Event Aggregator.",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Work your own magic here
		fmt.Println("start called")
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
