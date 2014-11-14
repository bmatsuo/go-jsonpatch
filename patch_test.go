package jsonpatch

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakePatch(t *testing.T) {
	// Test map
	docA := getMapDoc(`{"this":{"is":"my", "document":"sir"}}`)
	docB := getMapDoc(`{"this":{"document":"my", "is":"sir", "now":{"go":"away!"}}}`)
	patch, err := MakePatch(docA, docB)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(patch.Operations))
	// TODO: This test must depend on the ordering of a map iteration, because I'm getting
	// occasional failures occasionally with the ordering of the following two lines.
	assert.Equal(t, PatchOperation{Op: "replace", Path: "/this/is", Value: "sir"}, patch.Operations[0])
	assert.Equal(t, PatchOperation{Op: "replace", Path: "/this/document", Value: "my"}, patch.Operations[1])
	assert.Equal(t, PatchOperation{Op: "add", Path: "/this/now", Value: map[string]interface{}{"go": "away!"}}, patch.Operations[2])

	// Test array
	docA = getMapDoc(`{"a":[0, 1, 2, 3]}`)
	docB = getMapDoc(`{"a":[1, 2, 4, "hi"]}`)
	patch, err = MakePatch(docA, docB)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(patch.Operations))
	assert.Equal(t, PatchOperation{Op: "remove", Path: "/a/0"}, patch.Operations[0])
	assert.Equal(t, PatchOperation{Op: "replace", Path: "/a/2", Value: 4}, patch.Operations[1])
	assert.Equal(t, PatchOperation{Op: "add", Path: "/a/3", Value: "hi"}, patch.Operations[2])
}

func TestApplyPatchFromString(t *testing.T) {
	doc := getMapDoc(`{"foo": "bar"}`)

	patchOp, err := FromString(`[{"op": "add", "path": "/baz", "value": "qux"}]`)
	assert.Nil(t, err)
	patchOp.Apply(&doc)
	val, found := doc["baz"]
	assert.True(t, found)
	assert.Equal(t, "qux", val.(string))
}

// TestPatchMarshalJSON_empty ensures that an empty patch marshals as an empty
// array.
func TestPatchMarshalJSON_empty(t *testing.T) {
	js, err := json.Marshal(&Patch{nil})
	assert.Nil(t, err)
	assert.Equal(t, string(js), "[]")
}

// TestPatchJSON ensures that marshaled patches are unmarshaled to an identical
// patch.
func TestPatchJSON(t *testing.T) {
	for i, test := range []struct {
		p *Patch
	}{
		{&Patch{[]PatchOperation{}}},
		{&Patch{[]PatchOperation{
			{Op: Add, Path: "/foo", Value: "bar"},
		}}},
		{&Patch{[]PatchOperation{
			{Op: Remove, Path: "/foo"},
			{Op: Add, Path: "/bar", Value: "baz"},
		}}},
	} {
		var p *Patch
		index := fmt.Sprintf("test %d", i)
		js, err := json.Marshal(test.p)
		assert.Nil(t, err, index)
		err = json.Unmarshal(js, &p)
		assert.Nil(t, err, index)
		assert.Equal(t, p, test.p, index)
	}
}

func TestLcs(t *testing.T) {
	pairA, pairB := longestCommonSubseq(slice(1, 2, 3, 4), slice(0, 1, 2, 3, 5))
	assert.Equal(t, intPair{0, 3}, *pairA)
	assert.Equal(t, intPair{1, 4}, *pairB)

	pairA, pairB = longestCommonSubseq(slice(1, 3, 5), slice(0, 1, 2, 3, 4, 5, 6))
	assert.Equal(t, intPair{2, 3}, *pairA)
	assert.Equal(t, intPair{5, 6}, *pairB)
}

func TestSplitByCommonSeq(t *testing.T) {
	node := splitByCommonSeq(slice(0, 1, 2, 3), slice(1, 2, 4, 5), &intPair{0, -1}, &intPair{0, -1})
	assert.Nil(t, node.left)
	assert.Nil(t, node.right)

	// Left subtree
	assert.NotNil(t, node.leftPtr)
	assert.Equal(t, intPair{0, 1}, *node.leftPtr.left)
	assert.Nil(t, node.leftPtr.leftPtr)
	assert.Nil(t, node.leftPtr.rightPtr)
	// Right subtree
	assert.NotNil(t, node.rightPtr)
	assert.Equal(t, intPair{3, 4}, *node.rightPtr.left)
	assert.Equal(t, intPair{2, 4}, *node.rightPtr.right)
	assert.Nil(t, node.rightPtr.rightPtr)
	assert.Nil(t, node.rightPtr.leftPtr)
}

func slice(args ...interface{}) []interface{} {
	s := []interface{}{}
	s = append(s, args...)
	return s
}
