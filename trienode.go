package indexes

import (
	"gopkg.in/mgo.v2/bson"
)

// TrieNode defines a new TrieNode structure
type TrieNode struct {
	link  map[rune]*TrieNode
	IDSet *IDSet
}

/*
NewTrieNode returns a new nul Trie Node object
*/
func NewTrieNode() *TrieNode {
	return &TrieNode{make(map[rune]*TrieNode), NewIDSet()}
}

// GetLink will get the link at the specifed rune
func (tn *TrieNode) GetLink(r rune) *TrieNode {
	return tn.link[r]
}

// PutLink will place is a link for the given rune and link
func (tn *TrieNode) PutLink(r rune, link *TrieNode) {
	tn.link[r] = link
}

// GetAllRunes returns an array of all the keys in the map
func (tn *TrieNode) GetAllRunes() []rune {
	var keys []rune
	for k := range tn.link {
		keys = append(keys, k)
	}
	return keys
}

// RemoveLink returns an array of all the keys in the map
func (tn *TrieNode) RemoveLink(r rune) {
	tn.link[r] = nil
}

// SaveVal will save the passed in objectID into the TrieNode
func (tn *TrieNode) SaveVal(id bson.ObjectId) {
	tn.IDSet.SaveVal(id)
}

// GetVals will return the bson objectIDs for the node as an array
func (tn *TrieNode) GetVals() []bson.ObjectId {
	return tn.IDSet.GetVals()
}

// RemoveVal accepts an array of bson.objectIDs and sets the current node's value to this new array. Good for updating the node.
func (tn *TrieNode) RemoveVal(id bson.ObjectId) {
	tn.IDSet.Remove(id)
}

// ContainsVal returns true if the current node contains the given bson.objectID
func (tn *TrieNode) ContainsVal(id bson.ObjectId) bool {
	return tn.IDSet.ContainsVal(id)
}

// IsLeafNode returns bool true if the current node does not have any children links
// false is returned otherwise
func (tn *TrieNode) IsLeafNode() bool {
	return len(tn.GetAllRunes()) == 0
}
