package main

import (
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gocolly/colly/v2"
	"github.com/pkg/browser"
)

func generateCache() (hits []item) {
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
		e.ForEach("tr", func(_ int, el *colly.HTMLElement) {
			cname := el.Attr("class")
			if strings.HasPrefix(cname, "separator") || cname == "Heading" {
				return
			} else if strings.HasPrefix(cname, "memitem") {
				hits = append(hits,
					item{
						Title_:       fmt.Sprintf("%s %s", el.ChildText(".memItemLeft"), el.ChildText(".memItemRight")),
						Description_: "",
						URL:          fmt.Sprintf("%s%s", "https://llvm.org/doxygen/", el.ChildAttr(".memItemRight > a.el", "href"))})
				fmt.Println(hits[len(hits)-1])
			} else if strings.HasPrefix(cname, "memdesc") {
				hits[len(hits)-1].Description_ = el.ChildText(".mdescRight")
				fmt.Println(hits[len(hits)-1])
			}
		})
	})
	c.Visit("https://llvm.org/doxygen/group__LLVMC.html")
	return hits
}

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type item struct {
	Title_, Description_, URL string
}

func (i item) Title() string       { return i.Title_ }
func (i item) Description() string { return i.Description_ }
func (i item) FilterValue() string { return i.Title_ }

var _ list.Item = item{}

type model struct {
	list list.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, nil
		case "enter", "space":
			if !m.list.SettingFilter() && m.list.SelectedItem() != nil {
				err := browser.OpenURL(m.list.SelectedItem().(item).URL)
				if err != nil {
					fmt.Println("Couldn't open URL in browser: ")
					fmt.Println(err)
				}
			}
		}
	case tea.WindowSizeMsg:
		top, right, bottom, left := docStyle.GetMargin()
		m.list.SetSize(msg.Width-left-right, msg.Height-top-bottom)
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return docStyle.Render(m.list.View())
}

func main() {
	cachePath := os.ExpandEnv("$HOME/.local/share/llvm-c-search-cache.bin")
	var items []list.Item

	// Use saved search results when available, else crawl
	cacheFile, err := os.Open(cachePath)
	if errors.Is(err, os.ErrNotExist) {
		fmt.Printf("Search cache in '%s' does not exist. Generating.\n", cachePath)
		cacheFile, err = os.OpenFile(cachePath, os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			panic(err)
		}
		hits := generateCache()
		enc := gob.NewEncoder(cacheFile)
		err = enc.Encode(hits)
		if err != nil {
			panic(err)
		}
		for _, hit := range hits {
			items = append(items, hit)
		}
	} else {
		fmt.Printf("Using search cache.\n")
		fmt.Printf("To search again remove the '%s' file.\n", cachePath)
		enc := gob.NewDecoder(cacheFile)
		err = enc.Decode(&items)
		if err != nil {
			fmt.Printf("Error reading from cache in '%s'.\n", cachePath)
			fmt.Printf("Try deleting it to re-search.\n")
			os.Exit(1)
		}
	}
	m := model{list: list.NewModel(items, list.NewDefaultDelegate(), 0, 0)}
	m.list.Title = "LLVM C"

	p := tea.NewProgram(m)
	p.EnterAltScreen()

	if err := p.Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
