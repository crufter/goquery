// Package goquery gives you Jquery like functionality to scrape/manipulate html documents.
// The exp/html package taken from the experimental branch of the Go tree is provided to avoid installation hassle. It will be removed later, if
// it will become part of the standard.
package goquery

import (
	"bytes"
	"github.com/opesun/goquery/exp/html"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

type W struct{
}

func (w *W) Write(p []byte) (int, error) {
	fmt.Println(string(p))
	return len(p), nil
}

type Nodes []*Node

type Node struct {
	*html.Node
}

type Selector struct {
	Attributes map[string]interface{}
	Nodename   string
}

func getLevStr(lev int) string {
	str := ""
	for i := 0; i < lev; i++ {
		str += "   "
	}
	return str
}

func print(n *Node, lev int) {
	if n == nil {
		return
	}
	fmt.Println(getLevStr(lev), n.Data)
	for _, v := range n.Child {
		print(&Node{v}, lev+1)
	}
}

func (ns Nodes) Print() {
	for _, v := range ns {
		print(v, 0)
	}
}

func recur(n *Node, action func(*Node)) {
	action(n)
	for _, v := range n.Child {
		recur(&Node{v}, action)
	}
}

func recurStop(lev int, n *Node, action func(*Node, int) bool) {
	cont := action(n, lev)
	if cont {
		for _, v := range n.Child {
			recurStop(lev+1, &Node{v}, action)
		}
	}
}

func mapFromAttr(a []html.Attribute) map[string]string {
	m := map[string]string{}
	for _, v := range a {
		m[v.Key] = v.Val
	}
	return m
}

func mapFromSplit(a string) map[string]struct{} {
	m := map[string]struct{}{}
	sp := strings.Split(a, " ")
	for _, v := range sp {
		m[v] = struct{}{}
	}
	return m
}

// This is where we decide if a Node satisfies a given selector (for example: "div.nice" or "#whatever.yoyoyo")
func satisfiesSel(n *Node, sel Selector) bool {
	if len(sel.Nodename) > 0 {
		if sel.Nodename != n.Data {
			return false
		}
	}
	attr := mapFromAttr(n.Attr)
	for i, v := range sel.Attributes {
		if i == "class" {
			for _, k := range sel.Attributes["class"].([]string) {
				classm, has := attr["class"]
				if !has {
					return false
				}
				m := mapFromSplit(classm)
				_, hasc := m[k]
				if !hasc {
					return false
				}
			}
		} else {
			if val, ok := attr[i]; !ok || val != v {
				return false
			}
		}
	}
	return true
}

func findRecur(ns *Nodes, selector Selector) Nodes {
	sl := Nodes{}
	for _, v := range *ns {
		recur(v, func(n *Node) {
			if satisfiesSel(n, selector) {
				sl = append(sl, n)
			}
		})
	}
	return sl
}
func name(c *Selector, str string) {
	r, err := regexp.Compile("^([a-zA-Z0-9]*)")
	if err != nil {
		panic(err)
	}
	name := r.Find([]byte(str))
	if name != nil {
		c.Nodename = string(name)
	}
}

func class(c *Selector, str string) {
	r, err := regexp.Compile(`\.[a-zA-Z0-9\-]*`)
	if err != nil {
		panic(err)
	}
	classes := r.FindAll([]byte(str), -1)
	cl := []string{}
	for _, v := range classes {
		cl = append(cl, string(v)[1:])
	}
	if len(cl) > 0 {
		c.Attributes["class"] = cl
	}
}

func id(c *Selector, str string) {
	r, err := regexp.Compile(`\#[a-zA-Z0-9\-]*`)
	if err != nil {
		panic(err)
	}
	id := r.Find([]byte(str))
	if id != nil {
		c.Attributes["id"] = string(id)[1:]
	}
}

func attr(c *Selector, str string) {
	r, err := regexp.Compile(`\[[a-zA-Z0-9=]*\]`)
	if err != nil {
		panic(err)
	}
	attribs := r.FindAll([]byte(str), -1)
	for _, v := range attribs {
		to_split := string(v)[1:]
		to_split = to_split[:len(to_split)-1]
		att := strings.Split(to_split, "=")
		c.Attributes[att[0]] = att[1]
	}
}

func parseSingleSel(str string) Selector {
	c := Selector{Attributes: map[string]interface{}{}}
	name(&c, str)
	class(&c, str)
	id(&c, str)
	attr(&c, str)
	return c
}

func parseSelector(str string) []Selector {
	str_sl := strings.Split(str, " ")
	sel_sl := []Selector{}
	for _, v := range str_sl {
		sel_sl = append(sel_sl, parseSingleSel(v))
	}
	return sel_sl
}

func find(ns *Nodes, selector string) Nodes {
	sels := parseSelector(selector)
	if len(sels) < 2 {
		return findRecur(ns, sels[0])
	}
	first := sels[0]
	n := findRecur(ns, first)
	sels = sels[1:]
	ret := Nodes{}
	for _, v := range n {
		for _, j := range v.Child {
			recurStop(0, &Node{j}, func(no *Node, lev int) bool {
				if satisfiesSel(no, sels[lev]) {
					if lev == len(sels)-1 {
						ret = append(ret, no)
						return false
					}
					return true
				}
				return false
			})
		}
	}
	return ret
}

func (ns *Nodes) Find(selector string) Nodes {
	return find(ns, selector)
}

func Parse(htm string) (Nodes, error) {
	n, err := html.Parse(bytes.NewBufferString(htm) )
	return Nodes{&Node{n}}, err
}

func ParseUrl(ur string) (Nodes, error) {
	c := http.Client{}
	resp, err := c.Get(ur)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return Parse(string(body))
}
