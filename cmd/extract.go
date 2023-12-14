/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type extractCmdOpts struct {
	outputDir string
	compress  bool
	version   string
	tempDir   string
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
			checkErrorAndExit(extractOpts.fixOutputDirPath())
			tempDir, err := os.MkdirTemp(os.TempDir(), "microshift-install-*")
			checkErrorAndExit(err)
			defer func() {
				checkErrorAndExit(os.RemoveAll(tempDir))
			}()
			extractOpts.tempDir = tempDir + string(os.PathSeparator)
			checkErrorAndExit(extractManifestFiles(extractOpts.tempDir, args[0]))
			checkErrorAndExit(extractOpts.extractVersionFromTag(args[0]))
			checkErrorAndExit(extractOpts.createArchive(tempDir))
		},
	}
	extractCmd.PersistentFlags().StringVarP(&extractOpts.outputDir, "outputDir", "o", "/tmp", "output directory to place the extracted kustomize manifests as a tar file")
	extractCmd.PersistentFlags().BoolVarP(&extractOpts.compress, "compress", "c", false, "flag to indicate if the extracted manifest tar file needs to be compressed or not")
	return extractCmd
}

func (extractOpts *extractCmdOpts) createArchive(sourceDir string) error {
	// Create new Writers for gzip and tar
	// These writers are chained. Writing to the tar writer will
	// write to the gzip writer which in turn will write to
	// the file

	var tw *tar.Writer
	var tarFile *os.File
	var err error
	if extractOpts.compress {
		tarFile, err = os.Create(fmt.Sprintf("%sopenshift-gitops-microshift-bundle_%s.tar.gz", extractOpts.outputDir, extractOpts.version))
		if err != nil {
			return err
		}
		gw := gzip.NewWriter(tarFile)
		defer gw.Close()
		tw = tar.NewWriter(gw)
	} else {
		tarFile, err = os.Create(fmt.Sprintf("%sopenshift-gitops-microshift-bundle_%s.tar", extractOpts.outputDir, extractOpts.version))
		if err != nil {
			return err
		}
		tw = tar.NewWriter(tarFile)
	}
	defer tw.Close()
	fmt.Println("Created archive file " + tarFile.Name())

	files := []string{}
	filepath.WalkDir(sourceDir, func(path string, entry os.DirEntry, err error) error {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".yaml" {
			files = append(files, path)
		}
		return nil
	})

	// Iterate over files and add them to the tar archive
	for _, file := range files {
		err := extractOpts.addToArchive(tw, file)
		if err != nil {
			return err
		}
	}
	fmt.Println("Successfuly written archive file " + tarFile.Name() + " to archive")
	return nil
}

func (extractOpts *extractCmdOpts) addToArchive(tw *tar.Writer, filename string) error {
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

	header.Name = strings.Replace(filename, extractOpts.tempDir, "", 1)

	// Write file header to the tar archive
	err = tw.WriteHeader(header)
	if err != nil {
		return err
	}

	if !info.Mode().IsRegular() {
		//skip non-regular file
		return nil
	}

	fmt.Println("Writing file " + file.Name() + " to archive")
	// Copy file content to tar archive
	_, err = io.Copy(tw, file)
	if err != nil {
		return err
	}
	return nil
}

func (extractOpts *extractCmdOpts) fixOutputDirPath() error {
	// if the output dir does not end with path separator(/), then append the path separator
	if extractOpts.outputDir[len(extractOpts.outputDir)-1] != os.PathSeparator {
		extractOpts.outputDir = fmt.Sprintf("%s%c", extractOpts.outputDir, os.PathSeparator)
	}

	return createDirectory(extractOpts.outputDir)
}

func (extractOpts *extractCmdOpts) extractVersionFromTag(imageWithTag string) error {
	imageParts := strings.Split(imageWithTag, ":")
	if len(imageParts) >= 2 {
		extractOpts.version = imageParts[1]
	} else {
		fmt.Errorf("unable to get version tag from image url %s", imageWithTag)
	}
	return nil
}
