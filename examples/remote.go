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
	fmt.Println(x.Find("#eow-title").InnerHTML())
}
