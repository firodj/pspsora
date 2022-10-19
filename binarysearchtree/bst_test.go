package binarysearchtree

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type BSTStringTestSuite struct {
	suite.Suite
	bst *AVLTree[int, string]
}

func (st *BSTStringTestSuite) SetupTest() {
	st.bst = new(AVLTree[int, string])
	st.bst.Insert(80, "8")
	st.bst.Insert(40, "4")
	st.bst.Insert(100, "10")
	st.bst.Insert(20, "2")
	st.bst.Insert(60, "6")
	st.bst.Insert(10, "1")
	st.bst.Insert(30, "3")
	st.bst.Insert(50, "5")
	st.bst.Insert(70, "7")
	st.bst.Insert(90, "9")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestBSTStringTestSuite(t *testing.T) {
	suite.Run(t, new(BSTStringTestSuite))
}

func (st *BSTStringTestSuite) TestInsert() {
	st.bst.String()

	st.bst.Insert(11, "11")
	st.bst.String()

	st.Equal(11, st.bst.Size())
}

func (st *BSTStringTestSuite) TestInOrderTraverse() {
	var result []string
	st.bst.InOrderTraverse(func(i string) {
		result = append(result, i)
	})
	st.Equalf([]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}, result, "Traversal order incorrect, got %v", result)
}

func (st *BSTStringTestSuite) TestPreOrderTraverse() {
	var result []string
	st.bst.PreOrderTraverse(func(i string) {
		result = append(result, i)
	})
	st.Equalf([]string{"4", "2", "1", "3", "8", "6", "5", "7", "10", "9"}, result, "Traversal order incorrect, got %v instead of %v", result, []string{"8", "4", "2", "1", "3", "6", "5", "7", "10", "9", "11"})
}

func (st *BSTStringTestSuite) TestPostOrderTraverse() {
	var result []string
	st.bst.PostOrderTraverse(func(i string) {
		result = append(result, i)
	})
	st.Equalf([]string{"1", "3", "2", "5", "7", "6", "9", "10", "8", "4"}, result, "Traversal order incorrect, got %v instead of %v", result, []string{"1", "3", "2", "5", "7", "6", "4", "9", "11", "10", "8"})
}

func (st *BSTStringTestSuite) TestMin() {
	min := st.bst.Min()
	st.False(min.End(), "min not working")
	st.Equalf("1", min.Value(), "min should be 1")
}

func (st *BSTStringTestSuite) TestMax() {
	max := st.bst.Max()
	st.False(max.End(), "max not working")
	st.Equalf("10", max.Value(), "max should be 10")
}

func (st *BSTStringTestSuite) TestSearch() {
	it := st.bst.Search(10)
	st.False(it.End())
	st.Equal("1", it.Value())

	it = st.bst.Search(80)
	st.False(it.End())
	st.Equal("8", it.Value())

	it = st.bst.Search(110)
	st.True(it.End())
}

func (st *BSTStringTestSuite) TestRemove() {
	st.bst.Remove(10)
	st.Equal(9, st.bst.Size())

	min := st.bst.Min()
	st.False(min.End(), "min not working")
	st.Equalf("2", min.Value(), "min should be 2")

	st.bst.Remove(900)
	st.Equal(9, st.bst.Size())

	st.bst.Remove(60)
	st.bst.String()
	st.Equal(8, st.bst.Size())

	st.bst.Remove(40)
	st.bst.String()
	st.Equal(7, st.bst.Size())
}

func (st *BSTStringTestSuite) TestFloorCeil() {
	st.bst.Remove(10)

	var f, c Iterator[int, string]

	f, c = st.bst.FloorCeil(10)
	st.True(f.End())
	st.False(c.End())
	st.Equal(20, c.n.key)

	f, c = st.bst.FloorCeil(25)
	st.False(f.End())
	st.False(c.End())
	st.Equal(20, f.n.key)
	st.Equal(30, c.n.key)

	f, c = st.bst.FloorCeil(35)
	st.False(f.End())
	st.False(c.End())
	st.Equal(30, f.n.key)
	st.Equal(40, c.n.key)

	f, c = st.bst.FloorCeil(110)
	st.True(c.End())
	st.False(f.End())
	st.Equal(100, f.n.key)

	f, c = st.bst.FloorCeil(80)
	st.False(f.End())
	st.False(c.End())
	st.Equal(80, f.n.key)
	st.Equal(80, c.n.key)
}
