/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const (
	BUNDLE_MANIFESTS_FILE_NAME            = "openshift-gitops-microshift-bundle.tar"
	BUNDLE_MANIFESTS_COMPRESSED_FILE_NAME = "openshift-gitops-microshift-bundle.tar.gz"
)

type extractCmdOpts struct {
	outputDir string
	compress  bool
}

// extractCmd represents the extract command
func NewExtractCmd() *cobra.Command {
	extractOpts := &extractCmdOpts{}
	var extractCmd = &cobra.Command{
		Args:  cobra.ExactArgs(1),
		Use:   "extract",
		Short: "extracts the kustomize manifest files from the OLM bundle container to a tar file",
		Long:  `extracts the kustomize manifest files from the OLM bundle container to a tar file`,
		Run: func(cmd *cobra.Command, args []string) {
			err := extractManifestFiles(extractOpts.outputDir, args[0])
			if err != nil {
				panic(err)
			}
			err = extractOpts.createArchive()
			if err != nil {
				panic(err)
			}
		},
	}
	extractCmd.PersistentFlags().StringVarP(&extractOpts.outputDir, "outputDir", "o", "/tmp", "output directory to place the extracted kustomize manifests as a tar file")
	extractCmd.PersistentFlags().BoolVarP(&extractOpts.compress, "compress", "c", false, "flag to indicate if the extracted manifest tar file needs to be compressed or not")
	return extractCmd
}

func (extractOpts *extractCmdOpts) createArchive() error {
	// Create new Writers for gzip and tar
	// These writers are chained. Writing to the tar writer will
	// write to the gzip writer which in turn will write to
	// the file

	var tw *tar.Writer
	if extractOpts.compress {
		tarFile, err := os.Create(extractOpts.outputDir + string(os.PathSeparator) + BUNDLE_MANIFESTS_COMPRESSED_FILE_NAME)
		if err != nil {
			return err
		}
		gw := gzip.NewWriter(tarFile)
		defer gw.Close()
		tw = tar.NewWriter(gw)
	} else {
		tarFile, err := os.Create(extractOpts.outputDir + string(os.PathSeparator) + BUNDLE_MANIFESTS_FILE_NAME)
		if err != nil {
			return err
		}
		tw = tar.NewWriter(tarFile)
	}
	defer tw.Close()

	files := []string{}
	filepath.WalkDir(extractOpts.outputDir, func(path string, entry os.DirEntry, err error) error {
		if !entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tar.gz") {
			files = append(files, path)
		}
		return nil
	})

	// Iterate over files and add them to the tar archive
	for _, file := range files {
		err := addToArchive(tw, file)
		if err != nil {
			return err
		}
	}
	return nil
}

func addToArchive(tw *tar.Writer, filename string) error {
	// Open the file which will be written into the archive
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Get FileInfo about our file providing file size, mode, etc.
	info, err := file.Stat()
	if err != nil {
		return err
	}

	// Create a tar Header from the FileInfo data
	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		return err
	}

	// Use full path as name (FileInfoHeader only takes the basename)
	// If we don't do this the directory strucuture would
	// not be preserved
	// https://golang.org/src/archive/tar/common.go?#L626
	header.Name = filename

	// Write file header to the tar archive
	err = tw.WriteHeader(header)
	if err != nil {
		return err
	}

	// Copy file content to tar archive
	_, err = io.Copy(tw, file)
	if err != nil {
		return err
	}
	return nil
}
