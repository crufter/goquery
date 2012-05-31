package goquery

import(
	"fmt"
	"strings"
	"time"
	"net/http"
	"io/ioutil"
	"regexp"
)

type Nodes []*Node

func (ns Nodes) Print() {
	for _, v := range ns {
		v.Tree.print(v, 0)
	}
}

func (ns Nodes) HTML() []string {
	s := []string{}
	for _, v := range ns {
		s = append(s, v.Tree.Whole[v.Start:v.End])
	}
	return s
}

func (ns Nodes) InnerHTML() []string {
	s := []string{}
	for _, v := range ns {
		sub := v.Tree.Whole[v.Start:v.End]
		f := strings.Index(sub, ">")+1
		l := strings.LastIndex(sub, "<")
		s = append(s, sub[f:l])
	}
	return s
}

type Node struct {
	Tree		*Tree
	Parent		*Node
	Children	Nodes
	Attributes	map[string]interface{}
	Start, End	int
	Name		string
}

type Selector struct {
	Attributes 		map[string]interface{}
	Nodename		string
}

type Tree struct {
	Root 		*Node
	Whole 		string
	ParsingTook	time.Duration
}

		func getLevStr(lev int) string{
			str := ""
			for i:=0;i<lev;i++ {
				str += "   "
			}
			return str
		}
	
	func (t *Tree) print(n *Node, lev int) {
		if n == nil {
			return
		}
		fmt.Println(getLevStr(lev), n.Name)
		for _, v := range n.Children {
			t.print(v, lev+1)
		}
	}

func (t *Tree) Print() {
	t.print(t.Root, 0)
}

		func recur(n *Node, action func(*Node)) {
			action(n)
			for _, v := range n.Children {
				recur(v, action)
			}
		}
		
		func recurStop(lev int, n *Node, action func(*Node, int) bool) {
			cont := action(n, lev)
			if cont {
				for _, v := range n.Children {
					recurStop(lev+1, v, action)
				}
			}
		}
	
		// This is where we decide if a Node satisfies a given selector (for example: "div.nice" or "#whatever.yoyoyo")
		func satisfiesSel(n *Node, sel Selector) bool {
			if len(sel.Nodename) > 0 {
				if sel.Nodename != n.Name {
					return false
				}
			}
			for i, v := range sel.Attributes {
				if i == "class" {
					for _, k := range sel.Attributes["class"].([]string) {
						classm, has := n.Attributes["class"]
						if !has {
							return false
						}
						_, has = classm.(map[string]struct{})[k]
						if !has {
							return false
						}
					}
				} else {
					if val, ok := n.Attributes[i]; !ok || val != v {
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
					r, err := regexp.Compile("^([a-z0-9]*)")
					if err != nil {
						panic(err)
					}
					name := r.Find([]byte(str))
					if name != nil {
						c.Nodename = string(name)
					}
				}
			
				func class(c *Selector, str string) {
					r, err := regexp.Compile(`\.[a-z0-9\-]*`)
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
					r, err := regexp.Compile(`\#[a-z0-9\-]*`)
					if err != nil {
						panic(err)
					}
					id := r.Find([]byte(str))
					if id != nil {
						c.Attributes["id"] = string(id)[1:]
					}
				}
				
				func attr(c *Selector, str string) {
					r, err := regexp.Compile(`\[[a-z0-9=]*\]`)
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
			for _, j := range v.Children {
				recurStop(0, j, func(no *Node, lev int) bool {
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
	
func (t *Tree) Find(selector string) Nodes {
	return find(&Nodes{t.Root}, selector)
}

func (ns *Nodes) Find(selector string) Nodes {
	return find(ns, selector)
}

		func flip(b *bool) {
			if *b == false {
				*b = true
			} else {
				*b = false
			}
		}
	
		func reset(k, v *string, m map[string]interface{}) {
			m[*k] = *v
			*k = ""
			*v = ""
		}
	
		func adjustAttributes(m map[string]interface{}) {
			if v, ok := m["class"]; ok {
				clm := map[string]struct{}{}
				sl := strings.Split(v.(string), " ")
				for _, k := range sl {
					clm[k] = struct{}{}
				}
				m["class"] = clm
			}
		}
	
	// See * below
	func parseTag(tag string) (string, map[string]interface{}) {
		first_space := strings.Index(tag, " ")
		atts := map[string]interface{}{}
		var name string
		if first_space > 0 {
			name = tag[:first_space]
			inquote := false
			inval := false
			key, val := "", ""
			for i, v := range tag[first_space+1:] {
				sv := string(v)
				if (sv == `"` || sv == `'`) && string(tag[i-1]) != `\` {
					flip(&inquote)
					if !inquote {
						reset(&key, &val, atts)
					}
					continue
				}
				if sv == " " && inval && !inquote {
					if len(val) > 0 {	// Close val
						reset(&key, &val, atts)
						flip(&inval)
					}
					continue
				}
				if sv == "=" && !inval {
					flip(&inval)
					continue
				}
				
				if inval {
					val += sv
				} else {
					key += sv
				}
			}
			if len(key) > 0 {
				atts[key] = val
			}
		} else {
			name = tag
		}
		if len(atts) > 0 {
			adjustAttributes(atts)
		}
		return name, atts
	}

// Gotta write a proper parser later. *
func Parse(html string) *Tree {
	tim := time.Now()
	t := &Tree{}
	t.Whole = html
	pos := 0
	var parent *Node = t.Root
	for ;pos<len(html); {
		if len(html) > pos+2 && string(html[pos:pos+2]) == "</" {	// TODO: attribute value could contain </ < >...
			jump := strings.Index(html[pos+1:], ">")
			pos += jump - 1
			if parent != nil {
				parent.End = pos + jump
				if parent.Parent != nil {
					parent = parent.Parent
				}
			}
		} else if string(html[pos]) == "<" {	// TODO: attribute value could contain </ < >...
			jump := strings.Index(html[pos+1:], ">")
			name, atts := parseTag(html[pos+1:pos+1+jump])
			n := &Node{
				Tree:		t,
				Children: 	Nodes{},
				Name:		name,
				Parent:		parent,
				Start:		pos,
				Attributes:	atts,
			}
			if parent == nil {
				t.Root = n
			} else {
				parent.Children = append(parent.Children, n)
			}
			parent = n
			pos += jump
		}
		pos++
	}
	t.ParsingTook = time.Since(tim)
	return t
}

func ParseUrl(ur string) (*Tree, error) {
	c := http.Client{}
	resp, err := c.Get(ur)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return Parse(string(body)), nil
}