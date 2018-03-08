package crypto

import (
	."starchain/common"
	."starchain/errors"
	"errors"
	"bytes"
	"crypto/sha256"
)

type MerkleTreeNode struct{
	Hash Uint256
	Left *MerkleTreeNode
	Right *MerkleTreeNode

}

type MerkleTree struct{
	Depth uint
	Root *MerkleTreeNode
}


func ComputeRoot(hashes []Uint256) (Uint256,error){
	if len(hashes) == 0{
		return Uint256{},NewDetailErr(errors.New("input 0"),ErrNoCode,"")
	}
	if len(hashes) ==1 {
		return hashes[0],nil
	}
	tree,err := NewMerkleTree(hashes)
	if err != nil{
		return Uint256{},err
	}
	return tree.Root.Hash,nil
}



func (node *MerkleTreeNode) IsLeaf() bool{
	return node.Left == nil && node.Right == nil
}





func NewMerkleTree(hashes []Uint256)(*MerkleTree,error){
	if len(hashes) == 0{
		return nil,NewDetailErr(errors.New("newmerkletree input is nil"),ErrNoCode,"")
	}
	var height uint
	height = 1
	nodes := newLeafNodes(hashes)
	for len(nodes)>1{
		nodes = makeTopLevel(nodes)
		height += 1
	}
	mt := &MerkleTree{
		Depth:height,
		Root:nodes[0],
	}
	return mt,nil

}

func newLeafNodes(hashes []Uint256) []*MerkleTreeNode {
	var leaves []*MerkleTreeNode
	for _,data := range hashes{
		node := &MerkleTreeNode{
			Hash:data,
		}
		leaves = append(leaves,node)
	}
	return leaves
}

//calc a top level base on input
func makeTopLevel(nodes []*MerkleTreeNode) []*MerkleTreeNode{
	var topLevel []*MerkleTreeNode
	for i := 0; i < len(nodes)/2; i++ {
		topLevel = append(topLevel,joinNode(nodes[2*i],nodes[2*i+1]))
	}
	if len(nodes) % 2 ==1{
		index := len(nodes)-1
		topLevel = append(topLevel,joinNode(nodes[index],nodes[index]))
	}
	return topLevel
}

func joinNode(left,right *MerkleTreeNode) *MerkleTreeNode{
	var data []Uint256
	data = append(data,left.Hash)
	data = append(data,right.Hash)
	hash := DoubleHash(data)
	node := &MerkleTreeNode{
		Hash:hash,
		Left:left,
		Right:right,
	}
	return node
}

func DoubleHash(s []Uint256) Uint256{
	b := new(bytes.Buffer)
	for _, d := range s{
		d.Serialize(b)
	}
	temp := sha256.Sum256(b.Bytes())
	return Uint256(sha256.Sum256(temp[:]))
}