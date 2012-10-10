package main

import (
	"fmt"
	"github.com/mattbaird/goquery"
)

var example = `
<html>
	<head>
		<title>
		</title>
	</head>
</html>
<body>
	<div class=hey custom_attr="wow"><h2>Title here</h2></div>
	<span><h2>Yoyoyo</h2></span>
	<div id="x">
		<span>
			content<a href=""><div><li></li></div></a>
		</span>
	</div>
	<div class="yo hey">
		<a href="xyz"><div class="cow sheep bunny"><h8>content</h8></div></a>
	</div>
</body>
</html>
`

func main() {
	x, _ := goquery.Parse(example)
	x.Find("a div").Val("lol")
	fmt.Println(x.Html())
	fmt.Println(x.Find("div").HasClass("yo"))
	fmt.Println(x.Find("").Attrs("href"))
}
