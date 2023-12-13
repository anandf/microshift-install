/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the extract command
func NewVersionCmd() *cobra.Command {
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show the version",
		Long:  `Show the version`,
		Args:  cobra.ExactArgs(0),
		RunE:  runVersion,
	}
	return versionCmd
}

func runVersion(cmd *cobra.Command, args []string) error {
	fmt.Println("v0.0.1")
	return nil
}
