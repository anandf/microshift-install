/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"archive/tar"
	"bufio"
	"context"
	"os"

	"github.com/regclient/regclient"
	"github.com/regclient/regclient/types/ref"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "microshift-install",
	Short: "Used for installing OpenShift GitOps on a microshift instance",
	Long:  `Used for installing OpenShift GitOps on a microshift instance`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		client := regclient.New()
		f, err := os.Create("/tmp/microshift-bundle.tar")
		if err != nil {
			panic(err)
		}
		err = client.ImageExport(context.Background(), ref.Ref{Scheme: "reg", Registry: "registry-proxy.engineering.redhat.com", Repository: "rh-osbs/openshift-gitops-1-gitops-microshift-bundle", Digest: "sha256:4f12b3270db78ab7df567347003594edda9f6b0e7ba510b2a229c09a5b04fe90"}, bufio.NewWriter(f), regclient.ImageWithPlatforms([]string{"linux", "amd64"}))
		if err != nil {
			panic(err)
		}
		tarReader := tar.NewReader(f)
		tarReader.Next()

	},
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
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.microshift-install.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
