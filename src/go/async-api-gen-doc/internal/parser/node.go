package parser

import (
	"fmt"
	"strconv"
	"time"

	"github.com/dnitsch/async-api-generator/internal/gendoc"
	"github.com/dnitsch/async-api-generator/internal/token"
)

type GenDocBlock struct {
	Token        token.Token   `json:"beginToken"` // token.BEGIN_GEN_DOC token
	Annotation   gendoc.GenDoc `json:"annotation"`
	NodeCategory NodeCategory  `json:"docCategory"`
	Value        string        `json:"value"`
	EndToken     token.Token   `json:"endToken"`
}

// NodeCategory is an internal concept for assigning depth to the node
// the lower the number the higher the precedence
type NodeCategory int

const NodePrecedenceIncrement int = 0b0

const (
	_ NodeCategory = iota << NodePrecedenceIncrement
	ServiceNode
	ChannelNode
	OperationNode
	MessageNode
)

var nodeCatConverter = map[string]NodeCategory{
	"server":       ServiceNode,
	"root":         ServiceNode,
	"info":         ServiceNode,
	"channel":      ChannelNode,
	"subOperation": OperationNode,
	"pubOperation": OperationNode,
	"message":      MessageNode,
}

// GenDocTree structure includes an Index and the root node
type GenDocTree struct {
	Root  *GenDocNode
	Index map[string]*GenDocNode
}

const idxDivider string = "_#_"

func indexFromKey(key *GenDocNodeKey) string {
	return fmt.Sprintf("%v%s%s", key.Typ, idxDivider, key.Val)
}

func NewGenDocTree() *GenDocTree {
	root := NewGenDocNode(&GenDocBlock{}).WithKey(NewGenDocNodeKey(0, "root"))
	ntrie := &GenDocTree{
		Root:  root,
		Index: map[string]*GenDocNode{indexFromKey(root.key): root},
	}
	// add top level branches
	orphaned := NewGenDocNode(&GenDocBlock{}).WithKey(NewGenDocNodeKey(0, "orphaned"))
	parented := NewGenDocNode(&GenDocBlock{}).WithKey(NewGenDocNodeKey(0, "parented"))
	ntrie.AddNode(orphaned, ntrie.Root)
	ntrie.AddNode(parented, ntrie.Root)
	return ntrie
}

func (t *GenDocTree) ParentedBranch() *GenDocNode {
	return t.getNode(NewGenDocNodeKey(0, "parented"))
}

func (t *GenDocTree) OrhpanedBranch() *GenDocNode {
	return t.getNode(NewGenDocNodeKey(0, "orphaned"))
}

// GenDocNode base node for the n-ary doc tree
type GenDocNode struct {
	key      *GenDocNodeKey `json:"-"`
	Index    GenDocNodeKey  `json:"index"`
	Value    *GenDocBlock   `json:"value"`
	Children []*GenDocNode  `json:"children"`
	IsLeaf   bool           `json:"isLeaf"`
	// potentially add siblings
}

// GenDocNodeKey helper key for when a tree
//
//	can be loaded into a Radix and use the string as a key to walk it
type GenDocNodeKey struct {
	Typ NodeCategory `json:"typ"`
	Val string       `json:"val"`
}

func NewGenDocNodeKey(typ NodeCategory, val string) *GenDocNodeKey {
	return &GenDocNodeKey{typ, val}
}

func NewGenDocNode(value *GenDocBlock) *GenDocNode {
	return &GenDocNode{
		key:      &GenDocNodeKey{},
		Value:    value,
		Children: []*GenDocNode{},
	}
}

func (g *GenDocNode) WithKey(key *GenDocNodeKey) *GenDocNode {
	g.key = key
	// set value of key to index
	g.Index = *key
	return g
}

// IsKeyEqual compares the current node's key to the comparator Key
func (g *GenDocNode) IsKeyEqual(key *GenDocNodeKey) bool {
	return g.key.Typ == key.Typ && g.key.Val == key.Val
}

func (n *GenDocNode) addChild(child *GenDocNode) {
	n.Children = append(n.Children, child)
}

// SortLeafNodes returns all children of the parent either leaf or non-leaf nodes
//
// NOTE: this method uses a naked return
func (n *GenDocNode) SortLeafNodes() (leafs, nonleafs []*GenDocNode) {
	for _, node := range n.Children {
		if node.IsLeaf || len(node.Children) == 0 {
			leafs = append(leafs, node)
		} else {
			nonleafs = append(nonleafs, node)
		}
	}
	return
}

func (n *GenDocNode) IsRoot() bool {
	return n.key.Typ == 0 && n.key.Val == "root"
}

const LEAF_SUFFIX string = "__leaf__"

// preMergeId is a helper for _to be merged_ leaves for any given entity
func PreMergeKeyValSuffix(key *GenDocNodeKey) *GenDocNodeKey {
	// need to make a copy of the key
	k := *key
	// suffixing with randomness to not overwrite an existing index in the hashtable
	k.Val = k.Val + LEAF_SUFFIX + strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
	return &k
}

// AddNode inserts a node with the given value into the N-ary tree
func (t *GenDocTree) AddNode(node, parent *GenDocNode) {

	t.Index[indexFromKey(node.key)] = node

	if parent.IsRoot() {
		t.Root.addChild(node)
	} else {
		parent.addChild(node)
	}
}

// DeleteNode deletes a node with the given value from the N-ary tree
func (t *GenDocTree) DeleteNode(key *GenDocNodeKey) {
	delkey := indexFromKey(key)
	node := t.Index[delkey]
	if node == nil {
		return
	}

	delete(t.Index, delkey)
	// is root - deleting the entire tree
	if node.IsRoot() {
		t.Root = nil
		return
	}
	parent := t.FindParentNode(node)

	for i, child := range parent.Children {
		if child.IsKeyEqual(key) {
			parent.Children = append(parent.Children[:i], parent.Children[i+1:]...)
			break
		}
	}
}

// FindNode looks for a node in the tree by index
func (t *GenDocTree) FindNode(key *GenDocNodeKey) *GenDocNode {
	return t.getNode(key)
}

// getNode retrieves the node from a hashmap by key
func (t *GenDocTree) getNode(key *GenDocNodeKey) *GenDocNode {
	if node, ok := t.Index[indexFromKey(key)]; ok {
		return node
	}
	return nil
}

// FindParentNode does a BFS search
//
// need to do a level search in parented tree for existance of a node which satisfies the KeyLookup
//
// BFS should do here i.e. go through all the siblings on each level and look for a match
// the singleContext tree should be quite flat
func (t *GenDocTree) FindParentNode(chld *GenDocNode) *GenDocNode {
	// is root node
	if chld.IsRoot() {
		return nil
	}
	var found *GenDocNode = nil
	queue := []*GenDocNode{t.Root}
TreeLoop:
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		// look for a level with the required category
		// search all siblings on that level for the parentId
		// Enqueue children of each node
		for _, child := range node.Children {
			queue = append(queue, child)
			if child.IsKeyEqual(chld.key) {
				// found level with required category
				found = node
				break TreeLoop
			}
		}
	}
	return found
}
