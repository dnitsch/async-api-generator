package asyncapigendoc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dnitsch/async-api-generator/internal/generate"
	"github.com/dnitsch/async-api-generator/internal/parser"
	"github.com/dnitsch/async-api-generator/internal/storage"
	log "github.com/dnitsch/simplelog"
	"github.com/spf13/cobra"
)

var (
	Version  string = "0.0.1"
	Revision string = "1111aaaa"
)

var logger = log.New(os.Stderr, log.ErrorLvl)

var (
	outputStorageConfig        *storage.Conf
	inputLocationStorageConfig *storage.Conf
)

var (
	verbose        bool
	dryRun         bool
	outputLocation string
	inputLocation  string
)

var AsyncAPIGenCmd = &cobra.Command{
	Use:     "gendoc",
	Aliases: []string{"aadg", "generator"},
	Short:   "Generator for AsyncAPI documents",
	Long: `Generator for AsyncAPI documents, functions by performing lexical analysis on source files in a given base directory. 
These can then be further fed into other generator tools, e.g. client/server generators`,
	Example:      "",
	SilenceUsage: true,
	Version:      fmt.Sprintf("%s-%s", Version, Revision),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return setStorageLocation(inputLocation, outputLocation)
	},
}

func Execute(ctx context.Context) {
	if err := AsyncAPIGenCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func init() {
	AsyncAPIGenCmd.PersistentFlags().StringVarP(&outputLocation, "output", "o", "local://$HOME/.gendoc", `Output type and destination, currently only supports [local://, azblob://]. if dry-run is set then this is ignored`)
	AsyncAPIGenCmd.PersistentFlags().StringVarP(&inputLocation, "input", "i", "local://.", `Path to start the search in, Must include the protocol - see output for options`)
	AsyncAPIGenCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	AsyncAPIGenCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "", false, "Dry run only runs in validate mode and does not emit anything")
}

// config bootstraps pflags into useable config
//
// TODO: use viper
func config(outConf *storage.Conf) (*generate.Config, func(), error) {
	dirName := filepath.Base(outConf.Destination)

	conf := &generate.Config{
		ParserConfig:  parser.Config{ServiceRepoUrl: repoUrl, BusinessDomain: businessDomain, BoundedDomain: boundedCtxDomain, ServiceLanguage: repoLang},
		SearchDirName: dirName,
		Output:        outConf,
	}
	if isService {
		// use the current search dir name as the serviceId
		// this allows certain objects to __not__ have parentId or id specified
		conf.ParserConfig.ServiceId = dirName
	}

	if !dryRun {
		// create interim local dirs for interim state or interim download storage
		interim, err := os.MkdirTemp("", ".gendoc-interim-*")
		if err != nil {
			return nil, nil, err
		}
		download, err := os.MkdirTemp("", ".gendoc-download-*")
		if err != nil {
			return nil, nil, err
		}
		conf.InterimOutputDir = interim
		conf.DownloadDir = download
	}

	return conf, func() {
		_ = os.RemoveAll(conf.InterimOutputDir)
		_ = os.RemoveAll(conf.DownloadDir)
	}, nil
}

func setStorageLocation(input, output string) error {
	inStoreConf, err := storage.ParseStorageOutputConfig(input)
	if err != nil {
		return err
	}
	outStoreConf, err := storage.ParseStorageOutputConfig(output)
	if err != nil {
		return err
	}
	inputLocationStorageConfig = inStoreConf
	outputStorageConfig = outStoreConf
	return nil
}
