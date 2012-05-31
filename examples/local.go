package main

import(
	"fmt"
	"github.com/opesun/goquery"
	"time"
)

var example =
`<html>
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
		<a href="xyz"><div class="this and that"><h8>content</h8></div></a>
	</div>
</body>
</html>
`

func main() {
	x, _ := goquery.Parse(example)
	tim := time.Now()
	fmt.Println(x.Find("#eow-title").InnerHTML())
	fmt.Println(time.Since(tim))
	fmt.Println(x.ParsingTook)
	//z, err := ParseUrl("http://www.youtube.com/watch?v=b_TWgkAPqok")
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//fmt.Println(z.Find("h1").HTML())
	//fmt.Println(z.ParsingTook)
}