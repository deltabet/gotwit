package main

import (
    "fmt"
		"github.com/go-web-framework/templates"
		"os"
)

func main() {
		//do slices for fields?
		fmt.Printf("start\n")
    var s = templates.Set{}
		var x map[string]string = make(map[string]string)
		x["a"] = "b"
		x["c"] = "d"
		s.Parse("templates")
		err := s.Execute("test.html", os.Stdout, map[string]interface{} {
			"First": 2,
			"Second": "what",
			"Third": x,
		})
    if err != nil {
        fmt.Println(err)
    }
}
