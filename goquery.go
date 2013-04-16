// Package goquery gives you Jquery like functionality to [extract data from/manipulate] html documents.
// The exp/html package taken from the experimental branch of the Go tree is provided to avoid installation hassle. It will be removed later, if
// it will become part of the standard.
package goquery

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/opesun/goquery/exp/html"
)

type w struct {
	s string
}

func (wr *w) Write(p []byte) (int, error) {
	wr.s += string(p)
	return len(p), nil
}

type Nodes []*Node

type Node struct {
	*html.Node
}

type selector struct {
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

func fprint(w io.Writer, n *Node, lev int) {
	if n == nil {
		return
	}
	fmt.Fprintln(w, getLevStr(lev), n.Data)
	for _, v := range n.Child {
		fprint(w, &Node{v}, lev+1)
	}
}

func (ns Nodes) Fprint(w io.Writer) {
	for _, v := range ns {
		fprint(w, v, 0)
	}
}

func (ns Nodes) Print() {
	ns.Fprint(os.Stdout)
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
func satisfiesSel(n *Node, sel selector) bool {
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

func findRecur(ns *Nodes, selec selector) Nodes {
	sl := Nodes{}
	for _, v := range *ns {
		recur(v, func(n *Node) {
			if satisfiesSel(n, selec) {
				sl = append(sl, n)
			}
		})
	}
	return sl
}
func name(c *selector, str string) {
	r, err := regexp.Compile("^([a-zA-Z0-9]*)")
	if err != nil {
		panic(err)
	}
	name := r.Find([]byte(str))
	if name != nil {
		c.Nodename = string(name)
	}
}

func class(c *selector, str string) {
	r, err := regexp.Compile(`\.[a-zA-Z0-9\-\_]*`)
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

func id(c *selector, str string) {
	r, err := regexp.Compile(`\#[a-zA-Z0-9\-\_]*`)
	if err != nil {
		panic(err)
	}
	id := r.Find([]byte(str))
	if id != nil {
		c.Attributes["id"] = string(id)[1:]
	}
}

func attr(c *selector, str string) {
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

func parseSingleSel(str string) selector {
	c := selector{Attributes: map[string]interface{}{}}
	name(&c, str)
	class(&c, str)
	id(&c, str)
	attr(&c, str)
	return c
}

func parseSelector(str string) []selector {
	str_sl := strings.Split(str, " ")
	sel_sl := []selector{}
	for _, v := range str_sl {
		sel_sl = append(sel_sl, parseSingleSel(v))
	}
	return sel_sl
}

func find(ns *Nodes, selec string) Nodes {
	sels := parseSelector(selec)
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

// Parse a stream of html.
func Parse(r io.Reader) (Nodes, error) {
	n, err := html.Parse(r)
	return Nodes{&Node{n}}, err
}

// Parses a string which contains a html.
func ParseString(htm string) (Nodes, error) {
	return Parse(bytes.NewBufferString(htm))
}

// Parses a html document located at url.
func ParseUrl(ur string) (Nodes, error) {
	resp, err := http.Get(ur)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return Parse(resp.Body)
}

// html and htmlAll could be one func but heck I cba now.
func _html(ns Nodes, outer bool) string {
	if len(ns) == 0 {
		return ""
	}
	wr := w{}
	if outer {
		html.Render(&wr, ns[0].Node)
	} else {
		for _, v := range ns[0].Node.Child {
			html.Render(&wr, v)
		}
	}
	return wr.s
}

func htmlAll(ns Nodes, outer bool) []string {
	sl := []string{}
	for _, v := range ns {
		wr := w{}
		if outer {
			html.Render(&wr, v.Node)
		} else {
			for _, c := range v.Child {
				html.Render(&wr, c)
			}
		}
		sl = append(sl, wr.s)
	}
	return sl
}

func getAttr(ns Nodes, key string) string {
	if len(ns) == 0 {
		return ""
	}
	for _, v := range ns[0].Attr {
		if v.Key == key {
			return v.Val
		}
	}
	return ""
}

func getAttrs(ns Nodes, key string) []string {
	sl := []string{}
	for _, j := range ns {
		for _, v := range j.Attr {
			if v.Key == key {
				sl = append(sl, v.Val)
				break
			}
		}
	}
	return sl
}

func setAttrs(ns Nodes, key, val string) {
	for _, j := range ns {
		had := false
		for _, k := range j.Attr {
			if k.Key == key {
				k.Val = val
				had = true
			}
		}
		if !had {
			j.Attr = append(j.Attr, html.Attribute{"", key, val})
		}
	}
}

func removeAttr(ns Nodes, key string) Nodes {
	for _, v := range ns {
		for i, j := range v.Attr {
			if j.Key == key {
				v.Attr = append(v.Attr[:i], v.Attr[i+1:]...)
				break
			}
		}
	}
	return ns
}

// Reduce the set of matched elements to those that /*match the selector or*/ pass the function's test.
func filterByFunc(ns Nodes, f func(index int, element *Node) bool, negate bool) Nodes {
	sl := Nodes{}
	for i, v := range ns {
		verdict := f(i, v)
		if negate {
			verdict = !verdict
		}
		if verdict {
			sl = append(sl, v)
		}
	}
	return sl
}

// Reduce the set of matched elements to those that match the selector /*or pass the function's test*/.
func filterBySelector(ns Nodes, selector string, negate bool) Nodes {
	sel := parseSelector(selector)
	if len(sel) != 1 {
		return Nodes{}
	}
	return filterByFunc(ns,
		func(index int, e *Node) bool {
			return satisfiesSel(e, sel[0])
		}, negate)
}

func recurUp(ns Nodes, f func(e *Node)) {
	for _, v := range ns {
		f(v)
		if v.Parent != nil {
			recurUp(Nodes{&Node{v.Parent}}, f)
		}
	}
}

func recurUpBool(ns Nodes, f func(e *Node) bool) {
	for _, v := range ns {
		go_on := f(v)
		if go_on && v.Parent != nil {
			recurUpBool(Nodes{&Node{v.Parent}}, f)
		}
	}
}

//
// Here comes the JQuery-like API.
//

// Unfinished
// Adds the specified class(es) to each of the set of matched elements.
func (ns Nodes) AddClass(a string) {
}

// Get the value of an attribute for the first element in the set of matched elements.
// or
// Set one or more attributes for the set of matched elements.
func (ns Nodes) Attr(a ...string) string {
	if len(a) == 1 {
		return getAttr(ns, a[0])
	}
	setAttrs(ns, a[0], a[1])
	return ""
}

// Not in jQuery.
// Gets the value of an attribute for every element in a set.
func (ns Nodes) Attrs(key string) []string {
	return getAttrs(ns, key)
}

func (ns Nodes) Each(f func(index int, element *Node)) {
	for i, v := range ns {
		f(i, v)
	}
}

// Reduce the set of matched elements to the one at the specified index.
// index is an integer indicating the 0-based position of the element.
// -index is an integer indicating the position of the element, counting backwards from the last element in the set.
func (ns Nodes) Eq(index int) Nodes {
	l := len(ns)
	if index >= 0 {
		if index > l-1 {
			return Nodes{}
		}
		return Nodes{ns[index]}
	}
	if (l-1)+index < 0 {
		return Nodes{}
	}
	return Nodes{ns[(l-1)+index]}
}

// Reduce the set of matched elements to those that match the selector or pass the function's test.
func (ns Nodes) Filter(crit interface{}) Nodes {
	if s, ok := crit.(string); ok {
		return filterBySelector(ns, s, false)
	} else if f, ok := crit.(func(int, *Node) bool); ok {
		return filterByFunc(ns, f, false)
	}
	return Nodes{}
}

// Reduce the set of matched elements to those that match the selector or pass the function's test.
func (ns Nodes) First() Nodes {
	if len(ns) > 1 {
		return Nodes{ns[1]}
	}
	return ns
}

// Get the descendants of each element in the current set of matched elements, filtered by a selector, GoQuery object, or element.
func (ns *Nodes) Find(selector string) Nodes {
	return find(ns, selector)
}

// Check the current matched set of elements against a selector/*, element, or jQuery object*/ and return true if at least one of these elements matches the given arguments.
func (ns Nodes) Is(selector string) bool {
	sel := parseSelector(selector)
	if len(sel) == 1 {
		for _, v := range ns {
			if satisfiesSel(v, sel[0]) {
				return true
			}
		}
	}
	return false
}

// Reduce the set of matched elements to the final one in the set.
func (ns Nodes) Last() Nodes {
	l := len(ns)
	if l < 2 {
		return ns
	}
	return Nodes{ns[l-1]}
}

// Returns the number of elements in the GoQuery object.
func (ns Nodes) Length() int {
	return len(ns)
}

// Reduce the set of matched elements to those that have a descendant that matches the selector or DOM element.
func (ns Nodes) Has(selector string) Nodes {
	return ns.Filter(
		func(i int, e *Node) bool {
			has := false
			for _, v := range e.Child {
				n := Nodes{&Node{v}}
				if len(n.Find(selector)) > 0 {
					has = true
				}
			}
			return has
		})
}

// Determine whether any of the matched elements are assigned the given class.
func (ns Nodes) HasClass(cl string) bool {
	if len(strings.Split(cl, " ")) > 1 {
		return false
	}
	classes := getAttrs(ns, "class")
	for _, v := range classes {
		m := mapFromSplit(v)
		if _, ok := m[cl]; ok {
			return true
		}
	}
	return false
}

// Get the HTML contents of the first element in the set of matched elements.
func (ns Nodes) Html() string {
	return _html(ns, false)
}

// Not in jQuery.
// Get the HTML contents of all elements in the set of matched elements.
func (ns Nodes) HtmlAll() []string {
	return htmlAll(ns, false)
}

// Remove elements from the set of matched elements.
func (ns Nodes) Not(crit interface{}) Nodes {
	if s, ok := crit.(string); ok {
		return filterBySelector(ns, s, true)
	} else if f, ok := crit.(func(int, *Node) bool); ok {
		return filterByFunc(ns, f, true)
	}
	return Nodes{}
}

// Get the parent of each element in the current set of matched elements, optionally filtered by a selector.
func (ns Nodes) Parent(a ...string) Nodes {
	sl := Nodes{}
	sel := []selector{}
	if len(a) != 0 {
		sel = parseSelector(a[0])
		if len(sel) != 1 {
			return sl
		}
	}
	for _, v := range ns {
		if len(sel) == 0 {
			sl = append(sl, v)
		} else if satisfiesSel(&Node{v.Parent}, sel[0]) {
			sl = append(sl, v)
		}
	}
	return sl
}

// Get the ancestors of each element in the current set of matched elements, optionally filtered by a selector.
func (ns Nodes) Parents(a ...string) Nodes {
	sl := Nodes{}
	sel := []selector{}
	if len(a) != 0 {
		sel = parseSelector(a[0])
		if len(sel) != 1 {
			return sl
		}
	}
	for _, v := range ns {
		recurUp(Nodes{&Node{v.Parent}},
			func(e *Node) {
				if len(sel) == 0 {
					sl = append(sl, e)
				} else if satisfiesSel(&Node{e.Parent}, sel[0]) {
					sl = append(sl, e)
				}
			})
	}
	return sl
}

// Get the ancestors of each element in the current set of matched elements, up to but not including the element matched by the selector, DOM node, or jQuery object.
func (ns Nodes) ParentsUntil(a string) Nodes {
	sl := Nodes{}
	sel := parseSelector(a)
	if len(sel) == 0 {
		return sl
	}
	for _, v := range ns {
		recurUpBool(Nodes{&Node{v.Parent}},
			func(e *Node) bool {
				if satisfiesSel(&Node{e.Parent}, sel[0]) {
					return false
				}
				sl = append(sl, e)
				return true
			})
	}
	return sl
}

// Not in jQuery.
// Get the HTML contents of the first element in the set of matched elements, including the current element.
func (ns Nodes) OuterHtml() string {
	return _html(ns, true)
}

// Not in jQuery.
// Get the HTML contents of all elements in the set of matched elements, including the current elements.
func (ns Nodes) OuterHtmlAll() []string {
	return htmlAll(ns, true)
}

// Unfinished
// Remove the set of matched elements from the DOM.
func (ns Nodes) Remove() {

}

// Remove an attribute from each element in the set of matched elements.
func (ns Nodes) RemoveAttr(key string) Nodes {
	removeAttr(ns, key)
	return ns
}

// Reduce the set of matched elements to a subset specified by a range of indices.
func (ns Nodes) Slice(pos ...int) Nodes {
	plen := len(pos)
	l := len(ns)
	if plen == 1 && pos[0] < l-1 && pos[0] > 0 {
		return ns[pos[0]:]
	} else if len(pos) == 2 && pos[0] < l-1 && pos[1] < l-1 && pos[0] > 0 && pos[1] > 0 {
		return ns[pos[0]:pos[1]]
	}
	return Nodes{}
}

func text(buf *bytes.Buffer, n *Node) {
	if n == nil {
		return
	}
	if n.Type == html.TextNode {
		fmt.Fprintf(buf, "%v", n.Data)
	}
	for _, v := range n.Child {
		text(buf, &Node{v})
	}
}

// Get the combined text contents of each element in the set of matched elements, including their descendants.
func (ns Nodes) Text() string {
	buf := &bytes.Buffer{}

	for _, v := range ns {
		text(buf, v)
	}

	return buf.String()
}

// Get the current value of the first element in the set of matched elements.
// or
// Set the value of each element in the set of matched elements.
func (ns Nodes) Val(a ...string) string {
	l := len(a)
	if l == 0 {
		return getAttr(ns, "value")
	} else if l == 1 {
		setAttrs(ns, "value", a[0])
		return ""
	}
	return "Why more args than 1?" // Hehehe.
}
