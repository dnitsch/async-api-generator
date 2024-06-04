package asyncapigendoc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dnitsch/async-api-generator/internal/fshelper"
	"github.com/dnitsch/async-api-generator/internal/generate"
	"github.com/dnitsch/async-api-generator/internal/storage"
	log "github.com/dnitsch/simplelog"
	"github.com/spf13/cobra"
)

var (
	businessDomain   string
	boundedCtxDomain string
	repoUrl          string
	repoLang         string
	isService        bool
	serviceId        string
	singleCtxCmd     = &cobra.Command{
		Use:     "single-context",
		Aliases: []string{"sc", "single"},
		Short:   `Runs the gendoc against a single repo source`,
		Long:    `Runs the gendoc against a single repo source and emits the output to specified storage.`,
		RunE:    singleCtxExecute,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return setStorageLocation(inputLocation, outputLocation)
		},
	}
)

func init() {
	singleCtxCmd.PersistentFlags().StringVarP(&businessDomain, "business-domain", "b", "", `businessDomain e.g. Warehouse Systems`)
	singleCtxCmd.PersistentFlags().StringVarP(&boundedCtxDomain, "bounded-ctx", "c", "", `boundedCtxDomain`)
	singleCtxCmd.PersistentFlags().StringVarP(&repoUrl, "repo", "r", "", `repoUrl`)
	singleCtxCmd.PersistentFlags().StringVarP(&repoLang, "lang", "", "C#", `Main Language used in repo`)
	singleCtxCmd.PersistentFlags().StringVarP(&serviceId, "service-id", "", "", `serviceId`)
	singleCtxCmd.PersistentFlags().BoolVarP(&isService, "is-service", "s", false, `whether the repo is a service repo`)
	AsyncAPIGenCmd.AddCommand(singleCtxCmd)
}

func singleCtxExecute(cmd *cobra.Command, args []string) error {

	if verbose {
		logger = log.New(os.Stdout, log.DebugLvl)
	}

	conf, cleanUp, err := config(inputLocationStorageConfig)
	if err != nil {
		return err
	}

	defer cleanUp()

	files, err := fshelper.ListFiles(inputLocationStorageConfig.Destination)
	if err != nil {
		return err
	}

	gendoc := generate.New(conf, logger)

	gendoc.LoadInputsFromFiles(files)

	if err := gendoc.GenDocBlox(); err != nil {
		return err
	}

	if dryRun {
		logger.Debugf("--dry-run only not storing locally or remotely")
		return nil
	}

	// set out name for single repo analysis
	outName := fmt.Sprintf("current/%s.json", conf.SearchDirName)
	// select storage adapter
	sc, err := storage.ClientFactory(outputStorageConfig.Typ, outputStorageConfig.Destination)
	if err != nil {
		return err
	}

	storageUpldReq := &storage.StorageUploadRequest{ContainerName: outputStorageConfig.TopLevelFolder, BlobKey: outName, Destination: filepath.Join(outputStorageConfig.Destination, outputStorageConfig.TopLevelFolder, outName)}

	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	return gendoc.CommitInterimState(ctx, sc, storageUpldReq)
}
