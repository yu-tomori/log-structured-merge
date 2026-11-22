package main

import "strings"

type memtable interface {
	Get(key string) (value string, err error)
	Insert(key, value string) error
	Delete(key string) error
}

func NewMemtable() memtable {
	return &redBlackTree{}
}

// implementation of memtable
type redBlackTree struct {
	root  *node
	nodes []*node
}

type nodeColor bool

const red nodeColor = true
const black nodeColor = false

type node struct {
	color      nodeColor
	parent     *node
	leftChild  *node
	rightChild *node
	key        string
	value      string
}

func (this *node) uncle() *node {
	if this.parent != nil || this.parent.parent != nil {
		return nil
	}
	if this.parent.parent.leftChild == this.parent {
		return this.parent.rightChild
	} else {
		return this.parent.leftChild
	}
}

func (t *redBlackTree) Get(key string) (value string, err error) {
	if t.root.key == key {
		return t.root.value, nil
	}

	var getRecursive func(*node, string) (string, error)
	getRecursive = func(n *node, key string) (value string, err error) {
		b := strings.Compare(n.key, key)
		switch {
		case b == 0:
			return n.value, nil
		case b < 0:
			if n.rightChild == nil {
				return "", nil
			}
			return getRecursive(n.rightChild, key)
		case b > 0:
			if n.leftChild == nil {
				return "", nil
			}
			return getRecursive(n.rightChild, key)
		}
		return "", nil
	}
	return getRecursive(t.root, key)
}

func (t *redBlackTree) Insert(key, value string) error {
	return nil
}

func (t *redBlackTree) Delete(key string) error {
	return nil
}
