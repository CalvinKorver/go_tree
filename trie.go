package indexes

import (
	"strings"
	"sync"

	"gopkg.in/mgo.v2/bson"
)

// Trie defines a TrieIndex
type Trie struct {
	root *TrieNode
	mx   sync.RWMutex //RWMutex to protect the map
}

// NewTrie creates a new Trie object
func NewTrie() *Trie {
	return &Trie{
		root: NewTrieNode(),
	}
}

/*
Add constructs a tree of nodes based on the letters in the keys added to it. The tree starts with a single root node that holds no values. When a new key/value pair is added, the trie follows this algorithm:

let current node = root node
for each letter in the key
find the child node of current node associated with that letter
if there is no child node associated with that letter, create a new node and add it to current node as a child associated with the letter
set current node = child node
add value to current node
*/
func (t *Trie) Add(s string, id bson.ObjectId) *TrieNode {
	s = strings.ToLower(s)
	t.mx.Lock()
	curr := t.root
	for i := 0; i < len(s); i++ {
		r := rune(s[i])
		link := curr.GetLink(r)
		if link != nil {
			// If it contains an entry for our rune, we advance our search
			curr = curr.GetLink(r)
		} else {
			// It does not, so we need to create a new TrieNode there
			newNode := NewTrieNode()
			curr.PutLink(r, newNode)
			curr = newNode
		}
	}
	// We make sure that there isn't a duplicate id stored as a value already
	if !curr.ContainsVal(id) {
		curr.SaveVal(id)
	}
	t.mx.Unlock()
	return curr
}

/*
findTip helper function takes in a prefix and the currentNode to start the search. It traverses the Trie Index and stops when it reaches the last letter of the prefix and returns that TrieNode. If the prefix does not exist in the Trie, then it returns nil
*/
func findTip(prefix string, curr *TrieNode) *TrieNode {
	for i := 0; i < len(prefix); i++ {
		r := rune(prefix[i])
		if curr.GetLink(r) != nil {
			// If it contains an entry for our rune, we advance our search
			curr = curr.GetLink(r)
		} else {
			return nil
		}
	}
	return curr
}

/*
Remove handles the removal of a specific prefix/id pair from the Trie
Returns error if there is no prefix/id pair that exists in the Trie - nil otherwise
*/
func (t *Trie) Remove(prefix string, id bson.ObjectId) {
	removeHelper(t.root, prefix, id, 0)
}

func removeHelper(curr *TrieNode, prefix string, id bson.ObjectId, index int) bool {
	if index == len(prefix) {
		if !curr.ContainsVal(id) {
			return false
		}
		curr.RemoveVal(id)
		if (len(curr.GetVals())) == 0 && curr.IsLeafNode() {
			return true
		}
		return false
	}
	r := rune(prefix[index])
	node := curr.GetLink(r)
	if node == nil {
		return false
	}
	shouldDelete := removeHelper(node, prefix, id, (index + 1))
	if shouldDelete {
		curr.RemoveLink(r)
		return curr.IsLeafNode()
	}
	return false
}

// Get returns value if exists in the Trie index, otherwise nil
func (t *Trie) Get(prefix string) []bson.ObjectId {
	prefix = strings.ToLower(prefix)
	t.mx.RLock()
	defer t.mx.RUnlock()
	curr := findTip(prefix, t.root)
	if curr != nil {
		vals := curr.GetVals()
		if len(vals) != 0 {
			return vals
		}
	}
	return []bson.ObjectId{} //Empty
}

/*
GetMany gets the specified set of users
let current node = root node
for each letter in the prefix
	find the child node of current node associated with that letter
	if there is no child associated with that letter, no keys start with the prefix, so return and empty list
	set current node = child node
child node now points to the branch containing all keys that start with the prefix; recurse down the branch, gathering the keys and values, and return them
*/
func (t *Trie) GetMany(prefix string, n int) []bson.ObjectId {
	prefix = strings.ToLower(prefix)
	t.mx.RLock()
	defer t.mx.RUnlock()
	curr := findTip(prefix, t.root)
	res := NewIDSet()
	if curr != nil {
		depthFirst(curr, n, res)
		return res.GetVals()
	}
	return res.GetVals()
}

func depthFirst(curr *TrieNode, max int, res *IDSet) {
	if curr == nil {
		return
	}
	runes := curr.GetAllRunes()
	idList := curr.GetVals()

	if len(idList) > 0 { // There is a value(s) here
		for i := 0; i < len(idList); i++ {
			if res.Size() < max {
				res.SaveVal(idList[i])
			}
		}
		if curr.IsLeafNode() {
			return
		}
	}
	for _, currentRune := range runes {
		depthFirst(curr.GetLink(currentRune), max, res)
	}
}
