package main

import (
	"testing"
)

func rawMemtable(r rawNode, t *testing.T) memtable {
	rbt := &redBlackTree{}
	switch r.(type) {
	case blackN:
		rbt.root = &node{
			color:  black,
			parent: nil,
		}
	case redN:
		rbt.root = &node{
			color:  black,
			parent: nil,
		}
	default:
		t.Fatal()
	}
	rbt.root.key = r.key()
	rbt.root.value = r.value()
	setChild(rbt.root, r.left(), r.right())
	return rbt
}

func setChild(this *node, l, r rawNode) {
	if l != nil {
		switch l.(type) {
		case blackN:
			this.leftChild = &node{
				color: black,
			}
			setChild(this.leftChild, l.left(), l.right())
		case redN:
			this.leftChild = &node{
				color: red,
			}
		}
		this.leftChild.parent = this
		this.leftChild.key = l.key()
		this.leftChild.value = l.value()
		setChild(this.leftChild, l.left(), l.right())
	}

	if r != nil {
		switch r.(type) {
		case blackN:
			this.rightChild = &node{
				color: black,
			}
			setChild(this.rightChild, r.left(), r.right())
		case redN:
			this.rightChild = &node{
				color: red,
			}
		}
		this.rightChild.parent = this
		this.rightChild.key = r.key()
		this.rightChild.value = r.value()
		setChild(this.rightChild, r.left(), r.right())
	}
}

type rawNode interface {
	isNode()
	left() rawNode
	right() rawNode
	key() string
	value() string
}

type blackN struct {
	k, v string
	l, r rawNode
}

type redN struct {
	k, v string
	l, r rawNode
}

func (blackN) isNode()          {}
func (n blackN) left() rawNode  { return n.l }
func (n blackN) right() rawNode { return n.r }
func (n blackN) key() string    { return n.k }
func (n blackN) value() string  { return n.v }
func (redN) isNode()            {}
func (n redN) left() rawNode    { return n.l }
func (n redN) right() rawNode   { return n.r }
func (n redN) key() string      { return n.k }
func (n redN) value() string    { return n.v }

func TestInsert(t *testing.T) {
	m := NewMemtable()
	err := m.Insert("key1", "value1")
	if err != nil {
		t.Fatal(err)
	}
}

func TestMemtableGet(t *testing.T) {
	m := rawMemtable(
		blackN{
			k: "key5",
			v: "value5",
			l: blackN{
				k: "key4",
				v: "value4",
				l: blackN{
					k: "key3",
					v: "value3",
					l: nil,
					r: nil,
				},
				r: nil,
			},
			r: redN{
				k: "key10",
				v: "value10",
				l: blackN{
					k: "key8",
					v: "value8",
					l: nil,
					r: nil,
				},
				r: blackN{
					k: "key11",
					v: "value11",
					l: nil,
					r: nil,
				},
			},
		},
		t,
	)
	cases := []struct {
		title         string
		key, expected string
	}{
		{"root", "key5", "value5"},
		{"example", "key10", "value10"},
		{"not found", "key100", ""},
	}
	for _, c := range cases {
		actual, err := m.Get(c.key)
		if err != nil {
			t.Fatal(err)
		}
		if c.expected != actual {
			t.Errorf("%s:%s != %s", c.title, c.expected, actual)
		}
	}
}
