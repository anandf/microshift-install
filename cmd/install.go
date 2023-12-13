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
	"os/exec"
	"strings"

	"github.com/regclient/regclient"
	"github.com/regclient/regclient/types/manifest"
	"github.com/regclient/regclient/types/platform"
	"github.com/regclient/regclient/types/ref"
	"github.com/spf13/cobra"
)

type installCmdOpts struct {
	installDir                string
	postInstallRestartEnabled bool
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
			err := extractManifestFiles(installOpts.installDir, args[0])
			if err != nil {
				panic(err)
			}
			if installOpts.postInstallRestartEnabled {
				err := restartMicroShiftService()
				if err != nil {
					panic(err)
				}
			}
		},
	}
	installCmd.PersistentFlags().StringVarP(&installOpts.installDir, "installDir", "i", "/etc/microshift/", "directory to place the manifests extracted from the bundle container")
	installCmd.PersistentFlags().BoolVarP(&installOpts.postInstallRestartEnabled, "restart", "r", false, "post extraction of the manifests, whether the microshift service should be restarted")
	return installCmd
}

func extractManifestFiles(outputDir, imageRef string) error {
	ctx := context.Background()
	rc := regclient.New()
	//r := ref.Ref{Scheme: "reg", Registry: "registry-proxy.engineering.redhat.com", Repository: "rh-osbs/openshift-gitops-1-gitops-microshift-bundle", Tag: "v99.9.0-7"}
	r, err := ref.New(imageRef)
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
		if !strings.HasPrefix(header.Name, "manifests") || strings.HasSuffix(header.Name, "gitops-microshift-operator.clusterserviceversion.yaml") {
			fmt.Println("Skipping file:" + header.Name)
			continue
		}
		fileName := outputDir + string(os.PathSeparator) + header.Name
		switch header.Typeflag {
		case tar.TypeDir:
			_, err := os.Stat(fileName)
			if err != nil {
				if os.IsNotExist(err) {
					fmt.Println("Creating Directory:" + fileName)
					err := os.MkdirAll(fileName, 0775) // Create your file
					if err != nil {
						return err
					}
				}
			} else {
				fmt.Println("Directory " + fileName + " already present")
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

func restartMicroShiftService() error {
	fmt.Println("Restarting microshift service")
	cmd := exec.Command("systemctl", "restart", "microshift")
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
