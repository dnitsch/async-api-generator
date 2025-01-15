// Package generate is responsible for coordinating the
// reading of all files in a provided list.
//
// It will generate the required states - either interim or processed into AsyncAPI.
//
// It coordinates the writing any processed blocks or Trees into a remote storage as well.
//
// Example:
//
//	g := generate.New(...)
package generate

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	log "github.com/dnitsch/simplelog"

	"github.com/dnitsch/async-api-generator/internal/fshelper"
	"github.com/dnitsch/async-api-generator/internal/gendoc"
	"github.com/dnitsch/async-api-generator/internal/lexer"
	"github.com/dnitsch/async-api-generator/internal/parser"
	"github.com/dnitsch/async-api-generator/internal/storage"
	"github.com/dnitsch/async-api-generator/internal/token"
)

// Generate
type Generate struct {
	log       log.Loggeriface
	config    *Config
	inputs    []Input
	processed *Processed
	tree      *parser.GenDocTree
}

// Config holds the parser config
type Config struct {
	InterimOutputDir string // temp interim dir for storage of interim or processed files
	DownloadDir      string // temp dir for any remote downloads
	SearchDirName    string
	ParserConfig     parser.Config
	Output           *storage.Conf // nillable for now to optionally write to remote
	// Note: other properties can go here
	// perhaps better to use the options pattern
	// ...apply(opt)
}

type SchemaContent struct {
	EventId string // `json:"eventId"`
}

type SampleContent struct {
	EventId string // `json:"eventId"`
}

// Input contains additional info about the content to be lexed => parsed
// Such as the FileName, TODO: add more
type Input struct {
	FileName      string
	FullPath      string
	Content       string
	SchemaContent *SchemaContent
	SampleContent *SampleContent
}

// type FileList struct {
// 	Name string // fullname of file
// 	Path string // full path to file - either relative or full
// 	Type string // type of file e.g. schema json, CS, TF, K8sYaml, HelmYaml
// }

func New(config *Config, log log.Loggeriface) *Generate {
	return &Generate{
		config: config,
		log:    log,
		inputs: []Input{}, //init empty slice
	}
}

func (g *Generate) LoadInputsFromFiles(inputs []*fshelper.FileList) {

	for _, input := range inputs {
		b, _ := os.ReadFile(input.Path)
		i := Input{Content: string(b), FileName: input.Name, FullPath: input.Path}
		// crude condition to ensure we capture the contents of schema files if not defined inline
		// TODO: add schema files to orphaned >> schemas >> event_name
		if strings.Contains(input.Path, ".schema.json") {
			i.SchemaContent = &SchemaContent{EventId: strings.Split(input.Name, ".")[0]}
		}
		if strings.Contains(input.Path, ".sample.json") {
			i.SampleContent = &SampleContent{EventId: strings.Split(input.Name, ".")[0]}
		}
		g.inputs = append(g.inputs, i)
	}
}

// Processed holds sortable list of GenDocBlock
type Processed []parser.GenDocBlock

func (p Processed) Len() int      { return len(p) }
func (p Processed) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

// Less is a specific sorter func for Processed
// NodeCategory holds the key to precedence
func (p Processed) Less(i, j int) bool { return p[i].NodeCategory < p[j].NodeCategory }

// Processed returns Processed items
func (g *Generate) Processed() *Processed {
	return g.processed
}

type Generated struct {
	Source string
	GenDoc []parser.GenDocBlock
}

var ChunkSize int = 20

var ErrGenDocBlox = errors.New("GenDocBlox failed to parse all documents")

// GenDocBlox analyzes all the inputs and generates statements
// which are then used to create the precedent based tree
//
// Validation will be run at this point
func (g *Generate) GenDocBlox() error {
	prcsd := Processed{}
	parserConfig := g.config.ParserConfig
	errored := ""
	// parserChan used to hold result across goroutines
	// it can be encapsulated within this func
	type parserChan struct {
		err       error
		generated []parser.GenDocBlock
	}

	genCh := make(chan parserChan)

	wg := sync.WaitGroup{}

	// Implement some rate limiting in form semaphores
	// in case the source directory file count exceeds 50 files
	semaphoreChannel := make(chan struct{}, ChunkSize)
	defer close(semaphoreChannel)

	// process each file concurrently
	for idx, input := range g.inputs {
		wg.Add(1)
		semaphoreChannel <- struct{}{}
		go func(input Input, wg *sync.WaitGroup, idx int, sem chan struct{}) {
			defer wg.Done()
			if input.SchemaContent != nil {
				schemaBlock := []parser.GenDocBlock{
					{
						Token:        token.Token{Type: token.MESSAGE, Source: token.Source{File: input.FileName, Path: input.FullPath}, Literal: "", Line: 0, Column: 0},
						Value:        input.Content,
						NodeCategory: parser.MessageNode,
						Annotation: gendoc.GenDoc{
							Id: input.SchemaContent.EventId,
							// TODO: perhaps strip the version suffix from the message schema name
							// so that an operation parent can be found.
							// OR maybe we want to version the operations also
							// Parent:      input.SchemaContent.EventId,
							ContentType: gendoc.JSONSchema, CategoryType: gendoc.MessageBlock},
					},
				}
				// read from semaphore in case of schema file
				<-semaphoreChannel
				genCh <- parserChan{generated: schemaBlock}
				return
			}

			if input.SampleContent != nil {
				sampleBlock := []parser.GenDocBlock{
					{
						Token:        token.Token{Type: token.MESSAGE, Source: token.Source{File: input.FileName, Path: input.FullPath}, Literal: "", Line: 0, Column: 0},
						Value:        input.Content,
						NodeCategory: parser.MessageNode,
						Annotation: gendoc.GenDoc{
							Id: input.SampleContent.EventId,
							// TODO: perhaps strip the version suffix from the message schema name
							// so that an operation parent can be found.
							// OR maybe we want to version the operations also
							// Parent:      input.SchemaContent.EventId,
							ContentType: gendoc.Example, CategoryType: gendoc.MessageBlock},
					},
				}
				// read from semaphore in case of schema file
				<-semaphoreChannel
				genCh <- parserChan{generated: sampleBlock}
				return
			}

			l := lexer.New(lexer.Source{Input: input.Content, FileName: input.FileName, FullPath: input.FullPath})
			// pass in global config into the parser
			// with additional file/content info
			// as well as env info - e.g. name of service
			p := parser.New(l, &parserConfig).WithLogger(g.log)
			parsed, errs := p.InitialParse()
			if len(errs) > 0 {
				// read from semaphore in case of error
				<-semaphoreChannel
				genCh <- parserChan{err: fmt.Errorf("%v", errs)}
				return // trigger deffered
			}
			// read from semaphore
			<-sem
			genCh <- parserChan{generated: parsed}
		}(input, &wg, idx, semaphoreChannel)
	}

	// execute a non blocking waitGroup.Wait
	// close the genCh once complete
	go func() {
		wg.Wait()
		close(genCh)
	}()

	// range over a unbuffered channel
	// will contain [][]Generated{}
	for gc := range genCh {
		if gc.err != nil {
			errored += gc.err.Error() + "\n"
		}
		// append slice to existing slice
		prcsd = append(prcsd, gc.generated...)
	}

	if len(errored) > 0 {
		return fmt.Errorf("\n%s%w", errored, ErrGenDocBlox)
	}

	sort.Sort(prcsd)
	g.processed = &prcsd

	return nil
}

func (g *Generate) Tree() *parser.GenDocTree {
	return g.tree
}

// BuildContextTree
func (g *Generate) BuildContextTree() error {
	ntrie := parser.NewGenDocTree()
	g.tree = ntrie
	g.buildTree()

	g.secondPassExhaustive()

	return nil
}

func (g *Generate) buildTree() {

	for _, val := range *g.processed {
		v := val
		switch v.NodeCategory {
		// servers and infoblock IDs is the service itself and has no parent
		case parser.ServiceNode:
			assignServiceNode(g.tree, v)
		case parser.ChannelNode:
			assingParentedNode(g.tree, v, parser.ChannelNode)
		case parser.OperationNode:
			assingParentedNode(g.tree, v, parser.OperationNode)
		case parser.MessageNode:
			// TODO: message is a special case where a parent can also be looked up by channel
			// potential unparented messages canb belong to a channel
			assingParentedNode(g.tree, v, parser.MessageNode)
		}
	}
}

// assignServiceNode is a special case where we either find a node
// and assign a leaf of itself to be merged later
// or we create a new branch
func assignServiceNode(tree *parser.GenDocTree, v parser.GenDocBlock) {
	key := parser.NewGenDocNodeKey(parser.ServiceNode, v.Annotation.Id)
	leaf := parser.NewGenDocNode(&v).WithKey(parser.PreMergeKeyValSuffix(key))
	leaf.IsLeaf = true

	node := tree.FindNode(key)
	if node != nil {
		// node found adding only leaves to merge later
		tree.AddNode(leaf, node)
	} else {
		// node _NOT_ found need to add to root with
		// unmodified key + add leaf to new node
		parentedBranch := tree.FindNode(parser.NewGenDocNodeKey(0, "parented"))
		new := parser.NewGenDocNode(&v).WithKey(key)
		tree.AddNode(new, parentedBranch)
		// add leaf to that node for merging
		tree.AddNode(leaf, new)
	}
}

func assingParentedNode(tree *parser.GenDocTree, v parser.GenDocBlock, cat parser.NodeCategory) {
	key := parser.NewGenDocNodeKey(cat, v.Annotation.Id)

	// conceptual parent - e.g. a specific channel specified on a operation annotation
	parent := tree.FindNode(parser.NewGenDocNodeKey(key.Typ-1, v.Annotation.Parent))

	// create a leaf for later merging
	leaf := parser.NewGenDocNode(&v).WithKey(parser.PreMergeKeyValSuffix(key))
	leaf.IsLeaf = true
	if parent != nil {
		// parent is found
		cn := tree.FindNode(key)
		if cn != nil {
			// node already exists
			// add only leaf to sort out later via possible merging
			tree.AddNode(leaf, cn)
		} else {
			// node does not exists
			// add only leaf to sort out later via possible merging
			new := parser.NewGenDocNode(&v).WithKey(key)
			tree.AddNode(new, parent)
			// add child to appropriate  parent
			tree.AddNode(leaf, new)
		}
	} else {
		orpahnedBranch := tree.OrhpanedBranch()
		tree.AddNode(leaf, orpahnedBranch)
	}
}

// findOrphansNonLeafOwner looks for leaf node's owner inside the parented Tree
func (g *Generate) findOrphansNonLeafOwner() {
	// looking for parents for orphans
	// which may not have been assigned the first time around
	if len(g.tree.OrhpanedBranch().Children) > 0 {
		for _, orphanNode := range g.tree.OrhpanedBranch().Children {
			// assign to owner if exists
			nonLeafIndexVal := strings.Split(orphanNode.Index.Val, parser.LEAF_SUFFIX)
			nonleafOwner := g.tree.FindNode(parser.NewGenDocNodeKey(orphanNode.Index.Typ, nonLeafIndexVal[0]))
			if nonleafOwner != nil {
				node := parser.NewGenDocNode(orphanNode.Value).WithKey(&orphanNode.Index)
				g.tree.AddNode(node, nonleafOwner)
				// delete from orphaned branch tree
				g.tree.DeleteNode(&orphanNode.Index)
			}
		}
	}
}

func (g *Generate) findOrphansParents() {
	if len(g.tree.OrhpanedBranch().Children) > 0 {
		for _, orphanNode := range g.tree.OrhpanedBranch().Children {
			parent := g.tree.FindNode(parser.NewGenDocNodeKey(orphanNode.Index.Typ-1, orphanNode.Value.Annotation.Parent))
			if parent != nil {
				node := parser.NewGenDocNode(orphanNode.Value)
				g.tree.AddNode(node, parent)
				// delete from orphaned branch tree
				g.tree.DeleteNode(&orphanNode.Index)
			}
		}
	}
}

// secondPassExhaustive performs an exhaustive recursion
// going over orphans/and findparents _until_ there are
// no changes to the number of orphans in the tree
//
// There is _nothing_ wrong with orphans in the tree.
// Commonly caused by generated schemas/samples
func (g *Generate) secondPassExhaustive() {

	// if no orphans then we return straight away
	current, last, pass := len(g.tree.OrhpanedBranch().Children), 0, 0

	for current != last {
		pass += 1
		last = len(g.tree.OrhpanedBranch().Children)
		// Looking for NonLeafOwners
		// i.e. a summary or description or payload of an
		// operation/message/channel
		// at this point there shouldn't be too many unknown
		//
		g.findOrphansNonLeafOwner()

		// look for new parents
		// Note: maybe this goes before owner siblings
		g.findOrphansParents()
		current = len(g.tree.OrhpanedBranch().Children)
		g.log.Debugf("PASS: %d, found: %d orphans", pass, current)
	}
	g.log.Debugf("Exit after: %d passes", pass)
}

// TODO: message nodes should maybe go into a special pool...

// ConvertProcessed
func (g *Generate) ConvertProcessed() error {
	sortedProcessed := Processed{}
	for _, v := range g.inputs {
		gendocblox := new([]parser.GenDocBlock)
		if err := json.Unmarshal([]byte(v.Content), gendocblox); err != nil {
			return err
		}
		sortedProcessed = append(sortedProcessed, *gendocblox...)
	}
	sort.Sort(sortedProcessed)
	g.processed = &sortedProcessed
	return nil
}

func (g *Generate) AsyncAPIFromProcessedTree() error {
	orphans := g.Tree().OrhpanedBranch().Children
	if len(orphans) > 0 {
		for orphan := 0; orphan < len(orphans); orphan++ {
			g.log.Infof("_ORPHAN_ IdxType:%v IdxVal: %v", orphans[orphan].Index.Typ, orphans[orphan].Index.Val)
		}
	}
	serviceRoots := []*AsyncAPIRoot{}
	for _, node := range g.Tree().ParentedBranch().Children {
		cn := node
		asyncRoot, err := ConstructService(g.config, cn)
		if err != nil {
			return err
		}
		serviceRoots = append(serviceRoots, asyncRoot)
	}

	tp, err := NewTemplateProcessor()

	if err != nil {
		return err
	}
	return g.templateRoots(tp, serviceRoots)
}

func (g *Generate) templateRoots(tp TemplateProcessor, serviceRoots []*AsyncAPIRoot) error {
	for _, srvRoot := range serviceRoots {
		srv := srvRoot
		// generate new writer
		out := filepath.Join(g.config.InterimOutputDir, fmt.Sprintf("%s.yml", srv.ID))
		g.log.Debugf("writing file to: %s", out)
		f, err := os.Create(out)
		if err != nil {
			return err
		}

		if err := tp.GenerateFromRoot(f, *srv); err != nil {
			return err
		}
	}
	return nil
}

// CommitInterimState writes to disk (default or specified location) as well as remote storage
func (g *Generate) CommitInterimState(ctx context.Context, rc storage.StorageClient, rq *storage.StorageUploadRequest) error {

	b, err := json.Marshal(g.Processed())
	if err != nil {
		return err
	}
	rq.Reader = bytes.NewReader(b)
	return rc.Upload(ctx, rq)
}

// CommitProcessedState reads the emitted output from the global-context cmd into the InterimOutDir
//
// InterimOutput dir is a temporary location which is removed by the program on exit
func (g *Generate) CommitProcessedState(ctx context.Context, rc storage.StorageClient, rq *storage.StorageUploadRequest) error {

	prc, err := fshelper.ListFiles(g.config.InterimOutputDir)
	if err != nil {
		return err
	}

	// eg, ctx_ := errgroup.WithContext(ctx)
	// eg := new(errgroup.Group)

	for _, f := range prc {
		file := f
		req := *rq // copy requestObject
		// eg.Go(func() error {
		// 	file.Mu.Lock()
		// 	defer file.Mu.Unlock()
		// var reader io.Reader
		// reader = nil
		reader, err := os.OpenFile(file.Path, os.O_RDONLY, 0o500)
		if err != nil {
			return err
		}

		blobKey := fmt.Sprintf("asyncapi/%s", file.Name)
		req.BlobKey = blobKey
		dest := filepath.Join(req.Destination, req.ContainerName, blobKey)
		req.Destination = dest
		req.Reader = reader
		// })
		// return rc.Upload(ctx, req)
		if err := rc.Upload(ctx, &req); err != nil {
			return err
		}
	}
	// return eg.Wait()
	return nil
}
