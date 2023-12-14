/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type installCmdOpts struct {
	installDir                 string
	postInstallRestartDisabled bool
}

// installCmd represents the install command
func NewInstallCmd() *cobra.Command {
	installOpts := &installCmdOpts{}
	var installCmd = &cobra.Command{
		Args:  cobra.ExactArgs(1),
		Use:   "install",
		Short: "installs the kustomize based manifests in the microshift instance",
		Long:  `installs the kustomize based manifests in the microshift instance`,
		Run: func(cmd *cobra.Command, args []string) {
			checkErrorAndExit(installOpts.fixInstallDirPath())
			checkErrorAndExit(extractManifestFiles(installOpts.installDir, args[0]))
			if !installOpts.postInstallRestartDisabled {
				checkErrorAndExit(restartMicroShiftService())
			}
		},
	}
	installCmd.PersistentFlags().StringVarP(&installOpts.installDir, "installDir", "i", "/etc/microshift/manifests.d/", "directory to place the manifests extracted from the bundle container")
	installCmd.PersistentFlags().BoolVarP(&installOpts.postInstallRestartDisabled, "no-restart", "", false, "post extraction of the manifests, do not restart microshift service")
	return installCmd
}

func (installOpts *installCmdOpts) fixInstallDirPath() error {
	// if the output dir does not end with path separator(/), then append the path separator
	if installOpts.installDir[len(installOpts.installDir)-1] != os.PathSeparator {
		installOpts.installDir = fmt.Sprintf("%s%c", installOpts.installDir, os.PathSeparator)
	}
	return createDirectory(installOpts.installDir)
}
