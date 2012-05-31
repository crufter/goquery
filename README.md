goquery
=======

Jquery style selector engine for HTML documents, in Go.

Example
=======
See "remote.go" in the examples folder.

```
package main

import(
	"fmt"
	"github.com/opesun/goquery"
)

func main() {
	x, err := goquery.ParseUrl("http://www.youtube.com/watch?v=3-XxzRIyI_U&feature=related")
	if err != nil {
		panic(err)
	}
	fmt.Println(x.Find("#eow-title").InnerHTML())
}
```
This will output (if it can load the url):

```
[ <a id="watch-headline-show-title" href="/artist/Bounty_Killer?feature=watch_video_title">Bounty Killer</a> - Look ]
```