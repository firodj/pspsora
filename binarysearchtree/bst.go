package binarysearchtree

import (
	"fmt"
	"sync"
)

// Node a single node that composes the tree
type Node[T any] struct {
	key   int
	value T
	left  *Node[T] //left
	right *Node[T] //right
	parent *Node[T] // up
}

type Iterator[T any] struct {
	n *Node[T]
	// bst *ItemBinarySearchTree[T]
}

func (it *Iterator[T]) Value() T {
	return it.n.value
}

func (it *Iterator[T]) Key() int {
	return it.n.key
}

func (it *Iterator[T]) End() bool {
	return it.n == nil
}

func (it *Iterator[T]) Prev() Iterator[T] {
	return Iterator[T] {
		n: getPrevNode(it.n),
	}
}

func (it *Iterator[T]) Next() Iterator[T] {
	return Iterator[T] {
		n: getNextNode(it.n),
	}
}

// ItemBinarySearchTree the binary search tree of Items
type ItemBinarySearchTree[T any] struct {
	root *Node[T]
	lock sync.RWMutex
}

// Insert inserts the Item t in the tree
func (bst *ItemBinarySearchTree[T]) Insert(key int, value T) {
	bst.lock.Lock()
	defer bst.lock.Unlock()
	n := &Node[T]{key, value, nil, nil, nil}
	if bst.root == nil {
			bst.root = n
	} else {
			insertNode(bst.root, n)
	}
}

// internal function to find the correct place for a node in a tree
func insertNode[T any](node, newNode *Node[T]) {
	if newNode.key < node.key {
			if node.left == nil {
					node.left = newNode
					newNode.parent = node
			} else {
					insertNode(node.left, newNode)
			}
	} else {
			if node.right == nil {
					node.right = newNode
					newNode.parent = node
			} else {
					insertNode(node.right, newNode)
			}
	}
}

// InOrderTraverse visits all nodes with in-order traversing
func (bst *ItemBinarySearchTree[T]) InOrderTraverse(f func(T)) {
	bst.lock.RLock()
	defer bst.lock.RUnlock()
	inOrderTraverse(bst.root, f)
}

// internal recursive function to traverse in order
func inOrderTraverse[T any](n *Node[T], f func(T)) {
	if n != nil {
			inOrderTraverse(n.left, f)
			f(n.value)
			inOrderTraverse(n.right, f)
	}
}

// PreOrderTraverse visits all nodes with pre-order traversing
func (bst *ItemBinarySearchTree[T]) PreOrderTraverse(f func(T)) {
	bst.lock.Lock()
	defer bst.lock.Unlock()
	preOrderTraverse(bst.root, f)
}

// internal recursive function to traverse pre order
func preOrderTraverse[T any](n *Node[T], f func(T)) {
	if n != nil {
			f(n.value)
			preOrderTraverse(n.left, f)
			preOrderTraverse(n.right, f)
	}
}

// PostOrderTraverse visits all nodes with post-order traversing
func (bst *ItemBinarySearchTree[T]) PostOrderTraverse(f func(T)) {
	bst.lock.Lock()
	defer bst.lock.Unlock()
	postOrderTraverse(bst.root, f)
}

// internal recursive function to traverse post order
func postOrderTraverse[T any](n *Node[T], f func(T)) {
	if n != nil {
			postOrderTraverse(n.left, f)
			postOrderTraverse(n.right, f)
			f(n.value)
	}
}

// Min returns the Item with min value stored in the tree
func (bst *ItemBinarySearchTree[T]) Min() *T {
	bst.lock.RLock()
	defer bst.lock.RUnlock()
	n := bst.root
	if n == nil {
			return nil
	}
	for {
			if n.left == nil {
					return &n.value
			}
			n = n.left
	}
}

// Max returns the Item with max value stored in the tree
func (bst *ItemBinarySearchTree[T]) Max() *T {
	bst.lock.RLock()
	defer bst.lock.RUnlock()
	n := bst.root
	if n == nil {
			return nil
	}
	for {
			if n.right == nil {
					return &n.value
			}
			n = n.right
	}
}

// Search returns true if the Item t exists in the tree
func (bst *ItemBinarySearchTree[T]) Search(key int) Iterator[T] {
	bst.lock.RLock()
	defer bst.lock.RUnlock()
	return Iterator[T] {
		n: search(bst.root, key),
	}
}

// internal recursive function to search an item in the tree
func search[T any](n *Node[T], key int) *Node[T] {
	if n == nil {
			return nil
	}
	if key < n.key {
			return search(n.left, key)
	}
	if key > n.key {
			return search(n.right, key)
	}
	return n
}

// Remove removes the Item with key `key` from the tree
func (bst *ItemBinarySearchTree[T]) Remove(key int) {
	bst.lock.Lock()
	defer bst.lock.Unlock()
	remove(bst.root, key)
}

// internal recursive function to remove an item
func remove[T any](node *Node[T], key int) *Node[T] {
	if node == nil {
			return nil
	}
	if key < node.key {
			node.left = remove(node.left, key)
			node.left.parent = node
			return node
	}
	if key > node.key {
			node.right = remove(node.right, key)
			node.right.parent = node
			return node
	}
	// key == node.key
	if node.left == nil && node.right == nil {
			node = nil
			return nil
	}
	if node.left == nil {
			node = node.right
			return node
	}
	if node.right == nil {
			node = node.left
			return node
	}
	leftmostrightside := node.right
	for {
			//find smallest value on the right side
			if leftmostrightside != nil && leftmostrightside.left != nil {
					leftmostrightside = leftmostrightside.left
			} else {
					break
			}
	}
	node.key, node.value = leftmostrightside.key, leftmostrightside.value
	node.right = remove(node.right, node.key)
	node.right.parent = node
	return node
}

// String prints a visual representation of the tree
func (bst *ItemBinarySearchTree[T]) String() {
	bst.lock.Lock()
	defer bst.lock.Unlock()
	fmt.Println("------------------------------------------------")
	stringify(bst.root, 0)
	fmt.Println("------------------------------------------------")
}

// internal recursive function to print a tree
func stringify[T any](n *Node[T], level int) {
	if n != nil {
		format := ""
		for i := 0; i < level; i++ {
				format += "       "
		}
		format += "---[ "
		level++
		stringify(n.left, level)
		fmt.Printf(format+"%d\n", n.key)
		stringify(n.right, level)
	}
}

func (bst *ItemBinarySearchTree[T]) LowerBound(key int) Iterator[T] {
	bst.lock.RLock()
	defer bst.lock.RUnlock()

	return Iterator[T] {
		n: lowerBound(bst.root, key),
	}
}

// internal recursive function to lowerBound an item in the tree
func lowerBound[T any](n *Node[T], key int) *Node[T] {
	if n == nil {
		return nil
	}
	if key < n.key {
		if n.left == nil {
			return n
		}
		return lowerBound(n.left, key)
	}
	if key > n.key {
		return lowerBound(n.right, key)
	}
	return n
}

func getPrevNode[T any](node *Node[T]) *Node[T] {
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

func getNextNode[T any](node *Node[T]) *Node[T] {
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