package binarysearchtree

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type BSTStringTestSuite struct {
    suite.Suite
    bst ItemBinarySearchTree[string]
}

func (st *BSTStringTestSuite) SetupTest() {
    st.bst.Insert(8, "8")
    st.bst.Insert(4, "4")
    st.bst.Insert(10, "10")
    st.bst.Insert(2, "2")
    st.bst.Insert(6, "6")
    st.bst.Insert(1, "1")
    st.bst.Insert(3, "3")
    st.bst.Insert(5, "5")
    st.bst.Insert(7, "7")
    st.bst.Insert(9, "9")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func (st *BSTStringTestSuite) TestBSTStringTestSuite(t *testing.T) {
    suite.Run(t, new(BSTStringTestSuite))
}

func (st *BSTStringTestSuite) TestInsert() {
    st.bst.String()

    st.bst.Insert(11, "11")
    st.bst.String()
}

func (st *BSTStringTestSuite) TestInOrderTraverse() {
    var result []string
    st.bst.InOrderTraverse(func(i string) {
        result = append(result, i)
    })
    st.Equalf([]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11"}, result, "Traversal order incorrect, got %v", result)
}

func (st *BSTStringTestSuite) TestPreOrderTraverse() {
    var result []string
    st.bst.PreOrderTraverse(func(i string) {
        result = append(result, i)
    })
    st.Equalf([]string{"8", "4", "2", "1", "3", "6", "5", "7", "10", "9", "11"}, result, "Traversal order incorrect, got %v instead of %v", result, []string{"8", "4", "2", "1", "3", "6", "5", "7", "10", "9", "11"})
}

func (st *BSTStringTestSuite) TestPostOrderTraverse() {
    var result []string
    st.bst.PostOrderTraverse(func(i string) {
        result = append(result, i)
    })
    st.Equalf([]string{"1", "3", "2", "5", "7", "6", "4", "9", "11", "10", "8"}, result, "Traversal order incorrect, got %v instead of %v", result, []string{"1", "3", "2", "5", "7", "6", "4", "9", "11", "10", "8"})
}

func (st *BSTStringTestSuite) TestMin() {
    min := st.bst.Min()
    st.NotNil(min, "min not working")
    st.Equalf("1", *min, "min should be 1")
}

func (st *BSTStringTestSuite) TestMax() {
    max := st.bst.Max()
    st.NotNil(max, "max not working")
    st.Equalf("11", *max, "max should be 11")
}

func (st *BSTStringTestSuite) TestSearch() {
    it := st.bst.Search(1)
    st.False(it.End())
    st.Equal(1, it.Key())

    it = st.bst.Search(8)
    st.False(it.End())
    st.Equal(8, it.Key())

    it = st.bst.Search(11)
    st.False(it.End())
    st.Equal(11, it.Key())
}

func (st *BSTStringTestSuite) TestRemove() {
    st.bst.Remove(1)

    min := st.bst.Min()
    st.NotNil(min, "min not working")
    st.Equalf("2", *min, "min should be 2")
}

func (st *BSTStringTestSuite) TestLowerBound() {
    st.bst.Insert(20, "20")
    st.bst.Remove(1)
    st.bst.Insert(15, "15")

    it := st.bst.LowerBound(1)
    st.False(it.End())
    st.Equal(2, it.Key())

    it = it.Prev()
    st.True(it.End())

    it = it.Next()
    st.False(it.End())
    st.Equal(3, it.Key())

    it = st.bst.LowerBound(4)
    st.False(it.End())
    st.Equal(4, it.Key())

    it = it.Prev()
    st.False(it.End())
    st.Equal(3, it.Key())

    it = st.bst.LowerBound(12)
    st.False(it.End())
    st.Equal(15, it.Key())

    it = st.bst.LowerBound(19)
    st.False(it.End())
    st.Equal(20, it.Key())

    it = it.Prev()
    st.False(it.End())
    st.Equal(15, it.Key())

    it = st.bst.LowerBound(22)
    st.True(it.End())
}