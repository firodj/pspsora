package binarysearchtree

import (
	"fmt"
	"sync"

	"golang.org/x/exp/constraints"
)

// Node a single node that composes the tree
type Node[T any] struct {
	key   int
	value T
	left  *Node[T] //left
	right *Node[T] //right
	parent *Node[T] // up
	height int
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

	bst.root = insertNode(bst.root, key, value)
}

// internal function to find the correct place for a node in a tree
func insertNode[T any](node *Node[T], key int, value T) *Node[T] {
	if node == nil {
		return &Node[T]{key, value, nil, nil, nil, 1}
	} else if key < node.key {
		node.left = insertNode(node.left, key, value)
		node.left.parent = node
	} else {
		node.right = insertNode(node.right, key, value)
		node.right.parent = node
	}

	return rebalanceTreeAfterInsert(node, key)
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
func (bst *ItemBinarySearchTree[T]) Max() Iterator[T] {
	bst.lock.RLock()
	defer bst.lock.RUnlock()
	n := bst.root
	if n == nil {
			return Iterator[T]{nil}
	}
	for {
			if n.right == nil {
					return Iterator[T]{n}
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
	for n != nil && key != n.key {
		if key < n.key {
			n = n.left
		} else if key > n.key {
			n = n.right
		}
	}
	return n
}

// Remove removes the Item with key `key` from the tree
func (bst *ItemBinarySearchTree[T]) Remove(key int) {
	bst.lock.Lock()
	defer bst.lock.Unlock()

	bst.root = remove(bst.root, key)
}

// internal recursive function to remove an item
func remove[T any](node *Node[T], key int) *Node[T] {
	if node == nil {
			return nil
	} else if key < node.key {
			node.left = remove(node.left, key)
			if node.left != nil {
				node.left.parent = node
			}
	} else if key > node.key {
			node.right = remove(node.right, key)
			if node.right != nil {
				node.right.parent = node
			}
	} else  {
		// key == node.key
		//if node.left == nil && node.right == nil {
		//		node = nil
		//		return nil
		//} else
		if node.left == nil {
				parent := node.parent
				node = node.right
				if node != nil {
					node.parent = parent
				}
				return node
		} else
		if node.right == nil {
				parent := node.parent
				node = node.left
				if node != nil {
					node.parent = parent
				}
				return node
		} else {
			leftmostrightside := findSmallest(node.right)
			node.key, node.value = leftmostrightside.key, leftmostrightside.value

			node.right = remove(node.right, node.key)
			if node.right != nil {
				node.right.parent = node
			}
		}
	}

	return rebalanceTreeAfterRemove(node)
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
		p := ""
		if n.parent != nil {
			p = fmt.Sprintf("%d",n.parent.key)
		}
		fmt.Printf(format+"%d (%s)\n", n.key, p)
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
	for n != nil && n.key != key {
		if key < n.key {
			if n.left == nil || key > n.left.key {
				return n
			}
			n = n.left
		} else if key > n.key {
			n = n.right
		}
	}
	return n
}

func getPrevNode[T any](node *Node[T]) *Node[T] {
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

func getNextNode[T any](node *Node[T]) *Node[T] {
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

func getHeight[T any](n *Node[T]) int{
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

func recalculateHeight[T any](node *Node[T]) {
	node.height = 1 + max(getHeight(node.left), getHeight(node.right))
}

func getBalanceFactor[T any](node *Node[T]) int {
	if node == nil {
		return 0
	}
	return getHeight(node.left) - getHeight(node.right)
}

func findSmallest[T any](n *Node[T]) *Node[T] {
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

func rebalanceTreeAfterInsert[T any](node *Node[T], key int) *Node[T] {
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

func rebalanceTreeAfterRemove[T any](node *Node[T]) *Node[T] {
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


func rotateLeft[T any](n *Node[T]) *Node[T] {
	newRoot := n.right

	n.right = newRoot.left
	if n.right.parent != nil {
		n.right.parent = n
	}

	newRoot.left = n
	newRoot.parent = n.parent

	n.parent = newRoot

	recalculateHeight(n)
	recalculateHeight(newRoot)
	return newRoot
}

func rotateRight[T any](n *Node[T]) *Node[T] {
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