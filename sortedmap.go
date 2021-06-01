// Package gosortedmap implements sorted map with AVL tree
// api similar to buitin map
// Get, Set, Delete
// all operations are logn
// Iterator - AsChan, AsSlice
package gosortedmap

// implementation notes
// not very efficent (has bad const because of allocations and wrapping)
// delete algo also too big

// Comparable interface optionally can be implemented for object used as sorted map key
type Comparable interface {
	// CompareTo compares a and b and returns <0 when a < b, =0 when a == b, >0 when a > b
	CompareTo(interface{}) int
}

type comparableWrapper struct {
	value interface{}
	comp  Comparator
}

func (cw comparableWrapper) CompareTo(another interface{}) int {
	return cw.comp(cw.value, another)
}

// use callback instead, because it is handier

// Comparator compares a and b and returns <0 when a < b, =0 when a == b, >0 when a > b
type Comparator func(a, b interface{}) int

// SortedMap data structure
type SortedMap struct {
	tree *node
	comp Comparator
}

// Entry represents Key-Value pair
type Entry struct {
	Key   Comparable
	Value interface{}
}

func (sm *SortedMap) makeKeyComparable(key interface{}) Comparable {
	keycomp := Comparable(nil)
	if sm.comp != nil {
		keycomp = comparableWrapper{key, sm.comp}
	} else {
		keycomp = key.(Comparable)
	}
	return keycomp
}

// NewSortedMap creates new sorted map with comparator "comp"
// if comparator not specified, map will work if used keys implement Comparable
// else panics on operations
func NewSortedMap(comp Comparator) *SortedMap {
	sm := new(SortedMap)
	sm.comp = comp
	return sm
}

// Get value by key
func (sm *SortedMap) Get(key interface{}) (value interface{}, ok bool) {
	return find(sm.tree, sm.makeKeyComparable(key))
}

// Set value by key
func (sm *SortedMap) Set(key interface{}) {
	find(sm.tree, sm.makeKeyComparable(key))
}

// Delete value by key
func (sm *SortedMap) Delete(key interface{}) {
	sm.tree = remove(sm.tree, sm.makeKeyComparable(key))
}

// AsChan returns sorted Entries as channel
func (sm *SortedMap) AsChan() chan Entry {
	ch := make(chan Entry)
	go inOrderChan(sm.tree, ch)
	return ch
}

// AsSlice returns sorted Entries as slice
func (sm *SortedMap) AsSlice() []Entry {
	return inOrderSlice(sm.tree, nil)
}

type node struct {
	key         Comparable
	value       interface{}
	height      int
	left, right *node
}

func bfactor(n *node) int {
	if n == nil {
		return 0
	}
	return height(n.right) - height(n.left)
}

func height(n *node) int {
	if n == nil {
		return 0
	}
	return n.height
}

func fixHeight(n *node) {
	if n == nil {
		return
	}
	if height(n.left) > height(n.right) {
		n.height = 1 + height(n.left)
	} else {
		n.height = 1 + height(n.right)
	}
}

func rotateLeft(n *node) *node {
	m := n.right
	n.right, m.left = m.left, n
	fixHeight(n)
	fixHeight(m)
	return m
}

func rotateRight(n *node) *node {
	m := n.left
	n.left, m.right = m.right, n
	fixHeight(n)
	fixHeight(m)
	return m
}

func balance(n *node) *node {
	fixHeight(n)
	if bfactor(n) == 2 {
		if m := n.right; bfactor(m) < 0 {
			n.right = rotateRight(m)
		}
		return rotateLeft(n)
	}
	if bfactor(n) == -2 {
		if m := n.left; bfactor(m) > 0 {
			n.left = rotateLeft(m)
		}
		return rotateRight(n)
	}
	return n
}

func insert(n *node, k Comparable, v interface{}) *node {
	if n == nil {
		return &node{key: k, value: v, height: 1}
	}
	if k.CompareTo(n.key) < 0 {
		n.left = insert(n.left, k, v)
	} else if k.CompareTo(n.key) > 0 {
		n.right = insert(n.right, k, v)
	} else {
		n.value = v // update value
	}
	return balance(n)
}

func remove(n *node, k interface{}) *node {
	if n == nil {
		return n
	}
	t1, t2 := split(n, k, false)
	n = merge(t1, t2)
	return n
}

func findMin(n *node) *node {
	for n.left != nil {
		n = n.left
	}
	return n
}

func removeMin(n *node) *node {
	if n.left != nil {
		n.left = removeMin(n.left)
		return balance(n)
	}
	return n.right
}

func merge(n *node, m *node) *node {
	if n == nil {
		return m
	}
	if m == nil {
		return n
	}
	if n.height > m.height {
		n.right = merge(n.right, m)
		return balance(n)
	}
	if n.height+1 < m.height {
		m.left = merge(n, m.left)
		return balance(m)
	}
	r := findMin(m)
	m = removeMin(m)
	r.left, r.right = n, m
	return balance(r)
}

func inOrderSlice(n *node, res []Entry) []Entry {
	if n == nil {
		return res
	}
	res = inOrderSlice(n.left, res)
	res = append(res, Entry{n.key, n.value})
	res = inOrderSlice(n.right, res)
	return res
}

func inOrderChan(n *node, res chan Entry) {
	if n != nil {
		inOrderChan(n.left, res)
		res <- Entry{n.key, n.value}
		inOrderChan(n.right, res)
	}
}

func split(n *node, k interface{}, in bool) (t1 *node, t2 *node) {
	if n == nil {
		return nil, nil
	}
	if n.key.CompareTo(k) < 0 {
		t1, t2 = split(n.right, k, in)
		t1 = merge(n.left, t1)
		t1 = insert(t1, n.key, n.value)
	} else {
		t1, t2 = split(n.left, k, in)
		t2 = merge(t2, n.right)
		if n.key != k || in {
			t2 = insert(t2, n.key, n.value)
		}
	}
	return t1, t2
}

func find(n *node, k Comparable) (interface{}, bool) {
	for {
		if n == nil {
			return nil, false
		}
		if k.CompareTo(n.key) < 0 {
			n = n.left
		} else if k.CompareTo(n.key) > 0 {
			n = n.right
		} else {
			return n.value, true
		}
	}
}
