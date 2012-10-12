package main

import (
	"fmt"
	"github.com/opesun/goquery"
)

func main() {
	x, err := goquery.ParseUrl("http://www.youtube.com/watch?v=ob_nh1WMMzU")
	if err != nil {
		panic(err)
	}
	x.Find("#eow-title").Print()
	fmt.Println("---")
	x, err = goquery.ParseUrl("http://thepiratebay.se/search/one%20day%202011/0/99/0")
	if err != nil {
		panic(err)
	}
	x.Find("a.detLink").Print()
	fmt.Println("---")
	for _, v := range x.Find("a.detLink").HtmlAll() {
		fmt.Println(v)
	}
}
