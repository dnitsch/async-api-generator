package parser_test

import (
	"strings"
	"testing"

	"github.com/dnitsch/async-api-generator/internal/parser"
)

var testTree = func() *parser.GenDocTree {
	tree := parser.NewGenDocTree()
	// define nodes
	// root := parser.NewGenDocNode(&parser.GenDocBlock{})
	l1_a := parser.NewGenDocNode(&parser.GenDocBlock{}).WithKey(parser.NewGenDocNodeKey(parser.ServiceNode, "l1_a"))
	l1_b := parser.NewGenDocNode(&parser.GenDocBlock{}).WithKey(parser.NewGenDocNodeKey(parser.ServiceNode, "l1_b"))
	l1_c := parser.NewGenDocNode(&parser.GenDocBlock{}).WithKey(parser.NewGenDocNodeKey(parser.ServiceNode, "l1_c"))

	l2_a := parser.NewGenDocNode(&parser.GenDocBlock{}).WithKey(parser.NewGenDocNodeKey(parser.ChannelNode, "l2_a"))
	l2_b := parser.NewGenDocNode(&parser.GenDocBlock{}).WithKey(parser.NewGenDocNodeKey(parser.ChannelNode, "l2_b"))
	l2_c := parser.NewGenDocNode(&parser.GenDocBlock{}).WithKey(parser.NewGenDocNodeKey(parser.ChannelNode, "l2_c"))

	// make tree
	tree.AddNode(l1_a, tree.ParentedBranch())
	tree.AddNode(l1_b, tree.ParentedBranch())
	tree.AddNode(l1_c, tree.ParentedBranch())
	tree.AddNode(l2_a, l1_a)
	tree.AddNode(l2_b, l1_a)
	tree.AddNode(l2_c, l1_b)
	return tree
}

func Test_When_retrieving_nodes_via_either_DFS_BFS(t *testing.T) {
	tree := testTree()
	ttests := map[string]struct {
		root   *parser.GenDocNode
		lookup *parser.GenDocNodeKey
		expect *parser.GenDocNodeKey
	}{
		"when looking for existing it should return correct node": {
			root:   tree.Root,
			lookup: parser.NewGenDocNodeKey(parser.ServiceNode, "l1_a"),
			expect: parser.NewGenDocNodeKey(parser.ServiceNode, "l1_a"),
		},
		"when looking for nested existing it should return correct node": {
			root:   tree.Root,
			lookup: parser.NewGenDocNodeKey(parser.ChannelNode, "l2_c"),
			expect: parser.NewGenDocNodeKey(parser.ChannelNode, "l2_c"),
		},
		"when looking for non-existant node it should return nil": {
			root:   tree.Root,
			lookup: parser.NewGenDocNodeKey(parser.ServiceNode, "l2_z"),
			expect: nil,
		},
		"when passing in nil for root": {
			root:   tree.Root,
			lookup: parser.NewGenDocNodeKey(parser.ServiceNode, "l2_z"),
			expect: nil,
		},
		"when starting the tree traverse from a different node": {
			root:   tree.Root.Children[0],
			lookup: parser.NewGenDocNodeKey(parser.ChannelNode, "l2_a"),
			expect: parser.NewGenDocNodeKey(parser.ChannelNode, "l2_a"),
		},
	}
	for name, tt := range ttests {
		t.Run(name, func(t *testing.T) {

			gotBfs := tree.FindNode(tt.lookup)

			helperNodeLookUpValidator(t, "parser.FindNode", gotBfs, tt.expect)
		})
	}
}

func helperNodeLookUpValidator(t *testing.T, funcName string, got *parser.GenDocNode, expect *parser.GenDocNodeKey) {
	if got == nil {
		if expect != nil {
			t.Errorf("%s\nnot found assertion failed - got %v, wanted: %v", funcName, got, expect)
			return
		}
		return
	}
	if !got.IsKeyEqual(expect) {
		t.Fatalf("%s\nnot found node got: %v\nwanted: %v", funcName, got, expect)
	}
}

func Test_PreMergeKeyValSuffix_helper(t *testing.T) {
	ttests := map[string]struct {
		input  *parser.GenDocNodeKey
		expect *parser.GenDocNodeKey
	}{
		"simple string": {
			&parser.GenDocNodeKey{parser.ServiceNode, "bla"},
			&parser.GenDocNodeKey{parser.ServiceNode, "bla" + parser.LEAF_SUFFIX},
		},
		"special char . in the beginnin": {
			&parser.GenDocNodeKey{parser.ServiceNode, ".bar"},
			&parser.GenDocNodeKey{parser.ServiceNode, ".bar" + parser.LEAF_SUFFIX},
		},
		"special char #": {
			&parser.GenDocNodeKey{parser.ServiceNode, "#bar"},
			&parser.GenDocNodeKey{parser.ServiceNode, "#bar" + parser.LEAF_SUFFIX},
		},
	}
	for name, tt := range ttests {
		t.Run(name, func(t *testing.T) {
			got := parser.PreMergeKeyValSuffix(tt.input)
			if got.Typ != tt.expect.Typ {
				t.Errorf("Typ should match after merge, got: %v, wanted: %v", got.Typ, tt.expect.Typ)
			}
			ignoreRandSuffix := strings.Split(got.Val, "__")
			gs := strings.Join(ignoreRandSuffix[:2], "__")
			if gs+"__" != tt.expect.Val {
				t.Errorf("Val should match after merge, got: %v, wanted: %v", got.Val, tt.expect.Val)
			}
		})
	}
}

func Test_FindParentNode(t *testing.T) {

	tree := testTree()
	t.Run("OrhpanedBranch parent should be root", func(t *testing.T) {
		root := tree.FindParentNode(tree.OrhpanedBranch())
		if !root.IsRoot() {
			t.Errorf("found node is not root")
		}
	})

	t.Run("ParentedBranch parent should be root", func(t *testing.T) {
		root := tree.FindParentNode(tree.ParentedBranch())
		if !root.IsRoot() {
			t.Errorf("found node is not root")
		}
	})

	t.Run("root parent should be nil", func(t *testing.T) {
		root := tree.FindParentNode(tree.FindParentNode(tree.ParentedBranch()))
		if root != nil {
			t.Errorf("got: %v, wanted: <nil>", root)
		}
	})
}

func Test_DeleteNode(t *testing.T) {
	tree := testTree()
	t.Run("found node", func(t *testing.T) {
		tree.DeleteNode(parser.NewGenDocNodeKey(parser.ServiceNode, "l1_b"))
		found := tree.FindNode(parser.NewGenDocNodeKey(parser.ServiceNode, "l1_b"))
		if found != nil {
			t.Errorf("node not deleted properly, got: %v, wanted nil", found)
		}
	})

	t.Run("not found node no error", func(t *testing.T) {
		tree.DeleteNode(parser.NewGenDocNodeKey(parser.ServiceNode, "l1_z"))
		found := tree.FindNode(parser.NewGenDocNodeKey(parser.ServiceNode, "l1_z"))
		if found != nil {
			t.Errorf("node not deleted properly, got: %v, wanted <nil>", found)
		}
	})

	t.Run("deleting root successfully", func(t *testing.T) {
		tree.DeleteNode(parser.NewGenDocNodeKey(0, "root"))
		if tree.Root != nil {
			t.Errorf("node not deleted properly, got: %v, wanted <nil>", nil)
		}
	})
}

func Test_SortNodes(t *testing.T) {

	tree := testTree()
	ttests := map[string]struct {
		input        *parser.GenDocNode
		leafCount    int
		nonleafCount int
	}{
		"root should have 1 orphan leaf and 1 parented nonleaf": {
			tree.Root, 1, 1,
		},
		"orphaned should be empty of any nodes at this point": {
			tree.Root.Children[0], 0, 0,
		},
		"parented should only have nonleaf nodes": {
			tree.Root.Children[1], 1, 2,
		},
		"l1_a should have both nonleaf and leaf nodes": {
			tree.Root.Children[1].Children[0], 2, 0,
		},
	}
	for name, tt := range ttests {
		t.Run(name, func(t *testing.T) {
			l, nl := tt.input.SortLeafNodes()
			if len(l) != tt.leafCount {
				t.Errorf("wrong leaf count, got: %d, wanted: %d", len(l), tt.leafCount)
			}
			if len(nl) != tt.nonleafCount {
				t.Errorf("wrong nonleaf count, got: %d, wanted: %d", len(nl), tt.nonleafCount)
			}
		})
	}
}
