/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:           "microshift-install <cmd>",
		Short:         "Utility for installing OpenShift GitOps on a microshift instance",
		Long:          `Utility for installing OpenShift GitOps on a microshift instance`,
		SilenceUsage:  false,
		SilenceErrors: false,
	}

	rootCmd.AddCommand(
		NewVersionCmd(),
		NewInstallCmd(),
		NewExtractCmd(),
	)
	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := NewRootCmd().Execute()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
