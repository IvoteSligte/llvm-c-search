package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gocolly/colly/v2"
)

var cachePath = os.ExpandEnv("$HOME/.local/share/llvm-c-search-cache.txt")

func generateCache() (err error) {
	fmt.Printf("Generating search cache at '%s'.\n", cachePath)
	cacheFile, err := os.OpenFile(cachePath, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return
	}
	c := colly.NewCollector()
	c.OnHTML("table.memberdecls a.el", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if !(strings.HasPrefix(link, "group__") && strings.HasSuffix(link, ".html")) {
			return
		}
		fmt.Printf("%s\n%s\n", e.Text, strings.Repeat("-", len(e.Text)))
		c.Visit(e.Request.AbsoluteURL(link))
		fmt.Println()
	})
	c.OnHTML("table.memberdecls", func(e *colly.HTMLElement) {
		tableTitle := e.ChildText("tbody > .heading h2")
		if tableTitle != "Functions" {
			return
		}
		title := ""
		e.ForEach("tr", func(_ int, el *colly.HTMLElement) {
			cname := el.Attr("class")
			if strings.HasPrefix(cname, "separator") || cname == "Heading" {
				return
			} else if strings.HasPrefix(cname, "memitem") {
				if title != "" {
					fmt.Println(title)
					_, err = cacheFile.WriteString(title + "\n")
					if err != nil {
						return
					}
				}
				title = fmt.Sprintf("%s %s", el.ChildText(".memItemLeft"), el.ChildText(".memItemRight"))
			}
		})
		if title != "" {
			fmt.Println(title)
			_, err = cacheFile.WriteString(title + "\n")
			if err != nil {
				return
			}
		}
	})
	c.Visit("https://llvm.org/doxygen/group__LLVMC.html")
	err = cacheFile.Close()
	return
}

func main() {
	err := generateCache()
	if err != nil {
		panic(err)
	}
}
