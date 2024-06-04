package generate

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"text/template"

	sprig "github.com/Masterminds/sprig/v3"
	"github.com/dnitsch/async-api-generator/internal/gendoc"
	"github.com/dnitsch/async-api-generator/internal/parser"
)

// use embeded resources
const (
	templatesDir            = "templates"
	AsyncAPIRootCompleteTpl = "async-api-root-complete.yml"
)

var (
	//go:embed templates/*.yml
	templatefiles embed.FS
)

type TemplateProcessor struct {
	templates map[string]*template.Template
}

func NewTemplateProcessor() (TemplateProcessor, error) {
	d := TemplateProcessor{
		templates: make(map[string]*template.Template),
	}

	tmplFiles, err := fs.ReadDir(templatefiles, templatesDir)
	if err != nil {
		return d, err
	}

	for _, tmpl := range tmplFiles {
		if tmpl.IsDir() {
			continue
		}
		t := template.New(tmpl.Name()).Funcs(sprig.FuncMap())
		pt, err := t.ParseFS(templatefiles, templatesDir+"/"+tmpl.Name())
		if err != nil {
			return d, err
		}
		d.templates[tmpl.Name()] = pt
	}
	return d, nil
}

func (t TemplateProcessor) GenerateFromRoot(w io.Writer, input AsyncAPIRoot) error {

	foundTpl, ok := t.templates[AsyncAPIRootCompleteTpl]
	if !ok {
		return fmt.Errorf("not found the template specified")
	}

	return executeTemplate(w, foundTpl, input)
}

type docType interface {
	Operation | Server | Info | Channel | AsyncAPIRoot
}

func executeTemplate[T docType](w io.Writer, t *template.Template, data T) error {
	return t.Execute(w, data)
}

func ConstructService(conf *Config, srvNode *parser.GenDocNode) (*AsyncAPIRoot, error) {
	// first level services
	a := &AsyncAPIRoot{}
	a.AsyncAPI = "2.6.0" // include this value in the config
	a.DefaultContentType = "application/json"
	// leave this for now to automatically set to ID of service
	a.ID = srvNode.Value.Annotation.ServiceURN
	// set the title to be the ID - can be overwritten if title specifically set
	a.Info.Title = srvNode.Value.Annotation.Id
	a.Tags = append(a.Tags, []Tag{{Name: "repoUrl", Description: srvNode.Value.Annotation.ServiceRepoUrl}, {Name: "repoLang", Description: srvNode.Value.Annotation.ServiceRepoLang}}...)
	// get childleaf nodes
	srvMeta, channels := srvNode.SortLeafNodes()
	serviceConverter(srvMeta, a)
	a.Channels = map[string]Channel{}
ServiceLoop:
	for _, ch := range channels {
		chNode := ch
		chMeta, operations := chNode.SortLeafNodes()
		chann := &Channel{}
		channelConverter(chMeta, chann)
		currCh := chann
		oprtn := &Operation{}
		if len(operations) <= 0 {
			// assign known values to channel and continue to next channel
			a.Channels[chNode.Index.Val] = *currCh
			continue ServiceLoop
		}
		oprtn.OperationId = operations[0].Index.Val
		// there will only ever be a 1-2-1 mapping between operation -> messassage
		for _, op := range operations {
			opNode := op
			opMeta, messages := opNode.SortLeafNodes()
			operationConverter(opMeta, oprtn)
			// operation is either pub or sub
			switch opNode.Value.Annotation.CategoryType {
			case gendoc.PubOperationBlock:
				currCh.Publish = oprtn
			case gendoc.SubOperationBlock:
				currCh.Subscribe = oprtn
			}

			if len(messages) == 0 {
				oprtn.Message = nil
				// Do not add any message properties to this channel
				// if operation has no message Nodes
				//
				// This can be configurable and either skip adding the channel
				// completely or store incomplete on the service.
				// @david/@sree/@vincent one for the backlog
				// `Config` struct is already passed in so can be extended to dictate this behaviour
				a.Channels[chNode.Index.Val] = *currCh
				continue ServiceLoop
			}
			msgTop := &Message{}
			msgTop.MessageId = messages[0].Index.Val
			for _, msg := range messages {
				msg := msg
				// messages should only have leaf nodes
				msgMeta, _ := msg.SortLeafNodes()
				messageConverter(msgMeta, msgTop)
			}
			oprtn.Message = msgTop
		}
		a.Channels[chNode.Index.Val] = *currCh
	}

	return a, nil
}

func serviceConverter(nodes []*parser.GenDocNode, a *AsyncAPIRoot) {
	for _, srv := range nodes {
		switch srv.Value.Annotation.ContentType {
		case gendoc.Description:
			a.Info.Description = srv.Value.Value
		case gendoc.Title:
			// overwrite title if specifically set
			a.Info.Title = srv.Value.Value
		}
	}
}

func channelConverter(nodes []*parser.GenDocNode, ch *Channel) {
	for _, node := range nodes {
		switch node.Value.Annotation.ContentType {
		case gendoc.Description:
			ch.Description = node.Value.Value
		}
	}
}

func operationConverter(nodes []*parser.GenDocNode, op *Operation) {
	for _, node := range nodes {
		switch node.Value.Annotation.ContentType {
		case gendoc.Summary:
			op.Summary = node.Value.Value
		case gendoc.Description:
			op.Description = node.Value.Value
		}
	}
}

func messageConverter(nodes []*parser.GenDocNode, msg *Message) {
	for _, node := range nodes {
		switch node.Value.Annotation.ContentType {
		case gendoc.Summary:
			msg.Summary = node.Value.Value
		case gendoc.Description:
			msg.Description = node.Value.Value
		case gendoc.Title:
			msg.Title = node.Value.Value
		case gendoc.JSONSchema:
			msg.Payload = node.Value.Value
		case gendoc.Example:
			msg.Examples = append(msg.Examples, MessageBodyShared{
				// Name is set at a parsing level so will always be available here
				Name: msg.MessageId,
				Summary: fmt.Sprintf(`{"file":"%s[%d-%d]","path":"%s"}`,
					node.Value.Token.Source.File,
					node.Value.Token.Line, node.Value.EndToken.Line,
					node.Value.Token.Source.Path,
				),
				Payload: node.Value.Value,
			})
		}
	}
}

// TODO: explore generics approach for this
// func GenDocNodeConverter[T any](node *parser.GenDocNode, out T) (T, error) {
// 	fmt.Println(node)
// 	fmt.Println(out)
// 	return out, nil
// }
