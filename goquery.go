// Package goquery gives you Jquery like functionality to [extract data from/manipulate] html documents.
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
	s string
}

func (w *W) Write(p []byte) (int, error) {
	w.s += string(p)
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

// Parses a string which contains a html.
func Parse(htm string) (Nodes, error) {
	n, err := html.Parse(bytes.NewBufferString(htm) )
	return Nodes{&Node{n}}, err
}

// Parses a html document located at url.
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

// html and htmlAll could be one func but heck I cba now.
func _html(ns Nodes, outer bool) string {
	if len(ns) == 0 {
		return ""
	}
	w := W{}
	if outer {
		html.Render(&w, ns[0].Node)
	} else {
		for _, v := range ns[0].Node.Child {
			html.Render(&w, v)
		}
	}
	return w.s
}

func htmlAll(ns Nodes, outer bool) []string {
	sl := []string{}
	for _, v := range ns {
		w := W{}
		if outer {
			html.Render(&w, v.Node)
		} else {
			for _, c := range v.Child {
				html.Render(&w, c)
			}
		}
		sl = append(sl, w.s)
	}
	return sl
}

func getAttr(ns Nodes, key string) string {
	return ""
}

func setAttrs(ns Nodes, key, val string) {
	
}

//
// Here comes the JQuery-like API.
//

// Adds the specified class(es) to each of the set of matched elements.
func (ns Nodes) AddClass(a string) {
}


func (ns Nodes) Attr(key, val string) {

}

// Get the descendants of each element in the current set of matched elements, filtered by a selector, GoQuery object, or element.
func (ns *Nodes) Find(selector string) Nodes {
	return find(ns, selector)
}

// Returns the number of elements in the GoQuery object.
func (ns Nodes) Length() int {
	return len(ns)
}

// Get the HTML contents of the first element in the set of matched elements.
func (ns Nodes) Html() string {
	return _html(ns, false)
}

// Get the HTML contents of all elements in the set of matched elements.
func (ns Nodes) HtmlAll() []string {
	return htmlAll(ns, false)
}

// Get the parent of each element in the current set of matched elements, optionally filtered by a selector.
func (ns Nodes) Parent(a ...string) Nodes {
	return Nodes{}
}

// Get the ancestors of each element in the current set of matched elements, optionally filtered by a selector.
func (ns Nodes) Parents(a ...string) Nodes {
	return Nodes{}
}

// Get the ancestors of each element in the current set of matched elements, up to but not including the element matched by the selector, DOM node, or jQuery object.
func (ns Nodes) ParentsUntil(a ...string) Nodes {
	return Nodes{}
}

// Get the HTML contents of the first element in the set of matched elements, including the current element.
func (ns Nodes) OuterHtml() string {
	return _html(ns, true)
}

// Get the HTML contents of all elements in the set of matched elements, including the current elements.
func (ns Nodes) OuterHtmlAll() []string {
	return htmlAll(ns, true)
}


// Remove the set of matched elements from the DOM.
func (ns Nodes) Remove() {
}

// Remove an attribute from each element in the set of matched elements.
func (ns Nodes) RemoveAttr() {
}

// Get the current value of the first element in the set of matched elements.
// OR
// Set the value of each element in the set of matched elements.
func (ns Nodes) Val(a ...string) string {
	l := len(a)
	if l == 0 {
		return getAttr(ns, "value")
	} else if l == 1 {
		setAttrs(ns, "value", a[0])
		return ""
	}
	return "Why more args than 1?"	// Hehehe.
}
