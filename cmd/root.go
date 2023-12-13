/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/regclient/regclient"
	"github.com/regclient/regclient/types/manifest"
	"github.com/regclient/regclient/types/platform"
	"github.com/regclient/regclient/types/ref"
	"github.com/spf13/cobra"
)

type installCmd struct {
	outputDir                 string
	outputFormat              string
	postInstallRestartEnabled bool
}

func NewRootCmd() *cobra.Command {
	installOpts := &installCmd{}
	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Args:  cobra.ExactArgs(1),
		Use:   "microshift-install <image_ref>",
		Short: "Used for installing OpenShift GitOps on a microshift instance",
		Long:  `Used for installing OpenShift GitOps on a microshift instance`,
		// Uncomment the following line if your bare application
		// has an action associated with it:
		Run: func(cmd *cobra.Command, args []string) {
			err := installOpts.getFile(args)
			if err != nil {
				panic(err)
			}
		},
	}
	rootCmd.PersistentFlags().StringVarP(&installOpts.outputDir, "outputDir", "o", "/etc/microshift", "directory to place the manifests extracted from the bundle container")
	rootCmd.PersistentFlags().StringVarP(&installOpts.outputFormat, "outputFormat", "f", "tar.gz", "file format either tar or tar.gz of the extracted manifests")
	rootCmd.PersistentFlags().BoolVarP(&installOpts.postInstallRestartEnabled, "restart", "r", false, "post extraction of the manifests, whether the microshift service should be restarted")
	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := NewRootCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {}

func (installOpts *installCmd) getFile(args []string) error {
	ctx := context.Background()
	rc := regclient.New()
	//r := ref.Ref{Scheme: "reg", Registry: "registry-proxy.engineering.redhat.com", Repository: "rh-osbs/openshift-gitops-1-gitops-microshift-bundle", Tag: "v99.9.0-7"}
	r, err := ref.New(args[0])
	if err != nil {
		return err
	}
	m, err := rc.ManifestGet(ctx, r)
	if err != nil {
		return err
	}
	if m.IsList() {
		plat, err := platform.Parse("linux/amd64")
		if err != nil {
			return err
		}
		desc, err := manifest.GetPlatformDesc(m, &plat)
		if err != nil {
			return err
		}
		m, err = rc.ManifestGet(ctx, r, regclient.WithManifestDesc(*desc))
		if err != nil {
			return fmt.Errorf("failed to pull platform specific digest: %w", err)
		}
	}
	// Check if the manifest is of known image media type
	mi, ok := m.(manifest.Imager)
	if !ok {
		return fmt.Errorf("reference is not a known image media type")
	}
	layers, err := mi.GetLayers()
	if err != nil {
		return err
	}

	// Get the last layer which contains the manifests files
	blob, err := rc.BlobGet(ctx, r, layers[len(layers)-1])
	defer func() error {
		if err := blob.Close(); err != nil {
			return err
		}
		return nil
	}()
	if err != nil {
		return fmt.Errorf("failed pulling layer: %w", err)
	}
	btr, err := blob.ToTarReader()
	defer func() error {
		if err := btr.Close(); err != nil {
			return err
		}
		return nil
	}()

	if err != nil {
		return fmt.Errorf("could not convert layer to tar reader: %w", err)
	}
	treader, err := btr.GetTarReader()
	if err != nil {
		return err
	}
	for {
		header, err := treader.Next()
		if err == io.EOF {
			fmt.Println("Reached end of tar file")
			break
		}
		if err != nil {
			return err
		}
		if !strings.HasPrefix(header.Name, "manifests") {
			fmt.Println("Skipping file:" + header.Name)
			continue
		}
		fileName := installOpts.outputDir + header.Name
		switch header.Typeflag {
		case tar.TypeDir:
			fmt.Println("Creating Directory:" + fileName)
			if _, err := os.Stat(fileName); os.IsNotExist(err) {
				err := os.MkdirAll(fileName, 0775) // Create your file
				if err != nil {
					return err
				}
			} else {
				return err
			}
			continue
		case tar.TypeReg:
			fmt.Println("Creating file:" + fileName)
			var w io.Writer
			w, err = os.Create(fileName)
			if err != nil {
				return err
			}
			_, err = io.Copy(w, treader)
			if err != nil {
				return err
			}
		default:
			fmt.Printf("Unable to figure out type '%c' for file '%s'", header.Typeflag, header.Name)
		}
	}
	return nil
}
