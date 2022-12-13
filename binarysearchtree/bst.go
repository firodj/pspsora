package binarysearchtree

import (
	"fmt"
	"sync"

	"golang.org/x/exp/constraints"
)

// Node a single node that composes the tree
type Node[K constraints.Ordered, T any] struct {
	key    K
	value  T
	left   *Node[K, T] //left
	right  *Node[K, T] //right
	parent *Node[K, T] // up
	height int
}

type Iterator[K constraints.Ordered, T any] struct {
	n *Node[K, T]
}

func (it *Iterator[K, T]) Value() T {
	return it.n.value
}

func (it *Iterator[K, T]) Key() K {
	return it.n.key
}

func (it *Iterator[K, T]) End() bool {
	return it.n == nil
}

func (it *Iterator[K, T]) Prev() Iterator[K, T] {
	return Iterator[K, T]{
		n: getPrevNode(it.n),
	}
}

func (it *Iterator[K, T]) Next() Iterator[K, T] {
	return Iterator[K, T]{
		n: getNextNode(it.n),
	}
}

// AVLTree the binary search tree of Items
type AVLTree[K constraints.Ordered, T any] struct {
	root      *Node[K, T]
	lock      sync.RWMutex
	nodeCount int
}

func (bst *AVLTree[K, T]) Size() int {
	return bst.nodeCount
}

// Insert inserts the Item t in the tree
func (bst *AVLTree[K, T]) Insert(key K, value T) {
	bst.lock.Lock()
	defer bst.lock.Unlock()

	bst.root = bst.insertNode(bst.root, key, value)
}

func (bst *AVLTree[K, T]) createNode(key K, value T) *Node[K, T] {
	bst.nodeCount++
	return &Node[K, T]{key, value, nil, nil, nil, 1}
}

// internal function to find the correct place for a node in a tree
func (bst *AVLTree[K, T]) insertNode(node *Node[K, T], key K, value T) *Node[K, T] {
	if node == nil {
		return bst.createNode(key, value)
	} else if key < node.key {
		node.left = bst.insertNode(node.left, key, value)
		node.left.parent = node
	} else {
		node.right = bst.insertNode(node.right, key, value)
		node.right.parent = node
	}

	return rebalanceTreeAfterInsert(node, key)
}

// InOrderTraverse visits all nodes with in-order traversing
func (bst *AVLTree[K, T]) InOrderTraverse(f func(T)) {
	bst.lock.RLock()
	defer bst.lock.RUnlock()
	inOrderTraverse(bst.root, f)
}

// internal recursive function to traverse in order
func inOrderTraverse[K constraints.Ordered, T any](n *Node[K, T], f func(T)) {
	if n != nil {
		inOrderTraverse(n.left, f)
		f(n.value)
		inOrderTraverse(n.right, f)
	}
}

// PreOrderTraverse visits all nodes with pre-order traversing
func (bst *AVLTree[K, T]) PreOrderTraverse(f func(T)) {
	bst.lock.Lock()
	defer bst.lock.Unlock()
	preOrderTraverse(bst.root, f)
}

// internal recursive function to traverse pre order
func preOrderTraverse[K constraints.Ordered, T any](n *Node[K, T], f func(T)) {
	if n != nil {
		f(n.value)
		preOrderTraverse(n.left, f)
		preOrderTraverse(n.right, f)
	}
}

// PostOrderTraverse visits all nodes with post-order traversing
func (bst *AVLTree[K, T]) PostOrderTraverse(f func(T)) {
	bst.lock.Lock()
	defer bst.lock.Unlock()
	postOrderTraverse(bst.root, f)
}

// internal recursive function to traverse post order
func postOrderTraverse[K constraints.Ordered, T any](n *Node[K, T], f func(T)) {
	if n != nil {
		postOrderTraverse(n.left, f)
		postOrderTraverse(n.right, f)
		f(n.value)
	}
}

// Min returns the Item with min value stored in the tree
func (bst *AVLTree[K, T]) Min() Iterator[K, T] {
	bst.lock.RLock()
	defer bst.lock.RUnlock()
	n := bst.root
	if n == nil {
		return Iterator[K, T]{nil}
	}
	for {
		if n.left == nil {
			return Iterator[K, T]{n}
		}
		n = n.left
	}
}

// Max returns the Item with max value stored in the tree
func (bst *AVLTree[K, T]) Max() Iterator[K, T] {
	bst.lock.RLock()
	defer bst.lock.RUnlock()
	n := bst.root
	if n == nil {
		return Iterator[K, T]{nil}
	}
	for {
		if n.right == nil {
			return Iterator[K, T]{n}
		}
		n = n.right
	}
}

// Search returns true if the Item t exists in the tree
func (bst *AVLTree[K, T]) Search(key K) Iterator[K, T] {
	bst.lock.RLock()
	defer bst.lock.RUnlock()
	return Iterator[K, T]{
		n: search(bst.root, key),
	}
}

// internal recursive function to search an item in the tree
func search[K constraints.Ordered, T any](n *Node[K, T], key K) *Node[K, T] {
	if n == nil {
		return nil
	}
	for n != nil && key != n.key {
		if key < n.key {
			n = n.left
		} else if key > n.key {
			n = n.right
		}
	}
	return n
}

func (bst *AVLTree[K, T]) FloorCeil(key K) (floorIt, ceilIt Iterator[K, T]) {
	floor, ceil := floorCeil(bst.root, key)
	floorIt = Iterator[K, T]{floor}
	ceilIt = Iterator[K, T]{ceil}
	return
}

func floorCeil[K constraints.Ordered, T any](n *Node[K, T], key K) (floor, ceil *Node[K, T]) {
	for n != nil {
		if n.key == key {
			ceil = n
			floor = n
			break
		}

		if key > n.key {
			floor = n
			n = n.right
		} else {
			ceil = n
			n = n.left
		}
	}
	return
}

// Remove removes the Item with key `key` from the tree
func (bst *AVLTree[K, T]) Remove(key K) {
	bst.lock.Lock()
	defer bst.lock.Unlock()

	bst.root = bst.remove(bst.root, key)
}

func (bst *AVLTree[K, T]) freeNode(node *Node[K, T]) {
	if node != nil {
		bst.nodeCount -= 1
	}
}

// internal recursive function to remove an item
func (bst *AVLTree[K, T]) remove(node *Node[K, T], key K) *Node[K, T] {
	if node == nil {
		return node
	} else if key < node.key {
		node.left = bst.remove(node.left, key)
		if node.left != nil {
			node.left.parent = node
		}
	} else if key > node.key {
		node.right = bst.remove(node.right, key)
		if node.right != nil {
			node.right.parent = node
		}
	} else { // if key == node.key
		if node.left == nil {
			bst.freeNode(node)
			parent := node.parent
			node = node.right
			if node != nil {
				node.parent = parent
			}
			return node
		} else if node.right == nil {
			bst.freeNode(node)
			parent := node.parent
			node = node.left
			if node != nil {
				node.parent = parent
			}

			return node
		} else {
			leftmostrightside := findSmallest(node.right)
			node.key, node.value = leftmostrightside.key, leftmostrightside.value

			node.right = bst.remove(node.right, leftmostrightside.key)
			if node.right != nil {
				node.right.parent = node
			}
		}
	}

	return rebalanceTreeAfterRemove(node)
}

// String prints a visual representation of the tree
func (bst *AVLTree[K, T]) String() {
	bst.lock.Lock()
	defer bst.lock.Unlock()
	fmt.Println("------------------------------------------------")
	stringify(bst.root, 0)
	fmt.Println("------------------------------------------------")
}

// internal recursive function to print a tree
func stringify[K constraints.Ordered, T any](n *Node[K, T], level int) {
	if n != nil {
		format := ""
		for i := 0; i < level; i++ {
			format += "       "
		}
		format += "---[ "
		level++
		stringify(n.left, level)
		p := ""
		if n.parent != nil {
			p = fmt.Sprintf("%v", n.parent.key)
		}
		fmt.Printf(format+"%d (%s)\n", n.key, p)
		stringify(n.right, level)
	}
}

func getPrevNode[K constraints.Ordered, T any](node *Node[K, T]) *Node[K, T] {
	if node == nil {
		return nil
	}

	if node.left != nil {
		node = node.left
		for node.right != nil {
			node = node.right
		}
		return node
	}

	parent := node.parent
	for parent != nil && node == parent.left {
		node = parent
		parent = parent.parent
	}
	return parent
}

func getNextNode[K constraints.Ordered, T any](node *Node[K, T]) *Node[K, T] {
	if node == nil {
		return nil
	}

	if node.right != nil {
		node = node.right
		for node.left != nil {
			node = node.left
		}
		return node
	}

	parent := node.parent
	for parent != nil && node == parent.right {
		node = parent
		parent = parent.parent
	}
	return parent
}

func getHeight[K constraints.Ordered, T any](n *Node[K, T]) int {
	if n == nil {
		return 0
	}
	return n.height
}

func max[T constraints.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func recalculateHeight[K constraints.Ordered, T any](node *Node[K, T]) {
	node.height = 1 + max(getHeight(node.left), getHeight(node.right))
}

func getBalanceFactor[K constraints.Ordered, T any](node *Node[K, T]) int {
	if node == nil {
		return 0
	}
	return getHeight(node.left) - getHeight(node.right)
}

func findSmallest[K constraints.Ordered, T any](n *Node[K, T]) *Node[K, T] {
	leftmostrightside := n
	for {
		if leftmostrightside != nil && leftmostrightside.left != nil {
			leftmostrightside = leftmostrightside.left
		} else {
			break
		}
	}
	return leftmostrightside
}

func rebalanceTreeAfterInsert[K constraints.Ordered, T any](node *Node[K, T], key K) *Node[K, T] {
	if node == nil {
		return node
	}
	recalculateHeight(node)

	// check balance factor and rotateLeft if right-heavy and rotateRight if left-heavy
	balanceFactor := getBalanceFactor(node)
	if balanceFactor > 1 {
		// check if child is right-heavy and rotateLeft first
		if key < node.left.key {
			return rotateRight(node)
		} else {
			node.left = rotateLeft(node.left)
			return rotateRight(node)
		}
	}

	if balanceFactor < -1 {
		// check if child is left-heavy and rotateRight first
		if key > node.right.key {
			return rotateLeft(node)
		} else {
			node.right = rotateRight(node.right)
			return rotateLeft(node)
		}
	}
	return node
}

func rebalanceTreeAfterRemove[K constraints.Ordered, T any](node *Node[K, T]) *Node[K, T] {
	if node == nil {
		return node
	}
	recalculateHeight(node)

	// check balance factor and rotateLeft if right-heavy and rotateRight if left-heavy
	balanceFactor := getBalanceFactor(node)

	if balanceFactor > 1 {
		// check if child is right-heavy and rotateLeft first
		if getBalanceFactor(node.left) >= 0 {
			return rotateRight(node)
		} else {
			node.left = rotateLeft(node.left)
			return rotateRight(node)
		}
	}

	if balanceFactor < -1 {
		// check if child is left-heavy and rotateRight first
		if getBalanceFactor(node.right) <= 0 {
			return rotateLeft(node)
		} else {
			node.right = rotateRight(node.right)
			return rotateLeft(node)
		}
	}

	return node
}

func rotateLeft[K constraints.Ordered, T any](n *Node[K, T]) *Node[K, T] {
	newRoot := n.right

	n.right = newRoot.left
	if n.right != nil {
		n.right.parent = n
	}

	newRoot.left = n
	newRoot.parent = n.parent

	n.parent = newRoot

	recalculateHeight(n)
	recalculateHeight(newRoot)
	return newRoot
}

func rotateRight[K constraints.Ordered, T any](n *Node[K, T]) *Node[K, T] {
	newRoot := n.left

	n.left = newRoot.right
	if n.left != nil {
		n.left.parent = n
	}

	newRoot.right = n
	newRoot.parent = n.parent

	n.parent = newRoot

	recalculateHeight(n)
	recalculateHeight(newRoot)
	return newRoot
}
