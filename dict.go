package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"
)

var isSim = flag.Bool("sim", false, "find similar words")

const httpPath = "https://www.freedict.com/onldict/rus.html"
const tableSelector = "#right div table"
const fromWordsSelector = "tbody tr td:first-child"
const toWordsSelector = "tbody tr td:last-child"

type DictPair struct {
	Word        string
	Translation string
}

func (this DictPair) String() string {
	return fmt.Sprintf("%s -> %s", this.Word, this.Translation)
}

type TranslationDirection bool

const EnRu TranslationDirection = true
const RuEn TranslationDirection = false

func getDocument(word string, dir TranslationDirection, exact bool) *goquery.Document {
	var urlValues map[string][]string
	switch dir {
	case EnRu:
		urlValues = url.Values{
			"search": {word},
			"max":    {"50"},
			"dict":   {"eng2rus1"},
		}
	case RuEn:
		urlValues = url.Values{
			"search": {word},
			"max":    {"50"},
			"dict":   {"eng2rus2"},
		}
	}
	if exact {
		urlValues["exact"] = []string{"true"}
	}
	resp, err := http.PostForm(httpPath, urlValues)
	if err != nil {
		log.Fatal("Ooops,", err)
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		log.Fatal(err)
	}
	return doc
}

func getTranslations(word string, dir TranslationDirection, isExact bool) chan DictPair {
	resChan := make(chan DictPair, 50)
	go func() {
		document := getDocument(word, dir, isExact)
		table := document.Find(tableSelector).First()
		english_words := table.Find(fromWordsSelector)
		russian_words := table.Find(toWordsSelector)
		res := make([]DictPair, len(russian_words.Nodes))
		for i, engNode := range english_words.Nodes {
			res[i].Word = engNode.FirstChild.Data
			res[i].Translation = russian_words.Nodes[i].FirstChild.Data
		}
		for _, dictPair := range res {
			resChan <- dictPair
		}
		close(resChan)
	}()
	return resChan
}

func showHelp() {
	fmt.Println(`Format: dict [-sim] <word>
<word> - exactly one word
The -sim flag tells that you want to see the translation of 
this particular word, not looking for similar words.
`)
}
func main() {
	flag.Parse()
	if len(flag.Args()) != 1 {
		showHelp()
		return
	}
	word := flag.Args()[0]
	eng2ruChan := getTranslations(word, EnRu, !*isSim)
	ru2engChan := getTranslations(word, RuEn, !*isSim)
	for val := range eng2ruChan {
		fmt.Println(val)
	}
	for val := range ru2engChan {
		fmt.Println(val)
	}
}
