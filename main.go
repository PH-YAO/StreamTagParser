package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

func streamString(streamTagParser *StreamTagParser, inputStr string) {
	fmt.Printf("\n***   streaming tag parser input: %s   ***\n", inputStr)
	rand.Seed(time.Now().UnixNano())

	runes := []rune(inputStr) // convert to rune first to avoid breaking up multi-byte characters
	i := 0
	text := ""
	for i < len(runes) {
		// random length, 2 to 5 characters between
		step := rand.Intn(4) + 2
		if i+step > len(runes) {
			step = len(runes) - i
		}

		part := string(runes[i : i+step])

		tagTexts := streamTagParser.Parse(context.Background(), part)
		for _, tagText := range tagTexts {
			text += tagText.NormalText
			// output normal text before new tag text
			if tagText.NewTag && text != "" {
				fmt.Printf("previous text without TAG: %v\n", text)
				text = ""
			}
			// output normal text between tag scope
			if tagText.TagEnd {
				fmt.Printf("text: %v, tag: %v\n", text, tagText.TagText)
				text = ""
			}

		}
		i += step
	}
	// handle last normal text without tag
	if text != "" {
		fmt.Printf("last text out of tag: %v\n", text)
		text = ""
	}
}

func main() {
	// declare streamTagParser
	streamTagParser := NewStreamTagParser(context.Background(), "textTag")

	// test english parse
	streamString(streamTagParser, "Opening Ceremony<tag=Watch with a happy mood>A classic play is being performed.</tag>Closing Ceremony")
	// test chinese tag
	streamString(streamTagParser, "开幕式<tag=今天心情不错>这是一段话剧</tag>闭幕式")
	// test chinese and english text
	streamString(streamTagParser, "Opening Ceremony开幕式<tag=Watch with a happy mood今天心情不错>A classic play is being performed.这是一段话剧</tag>Closing Ceremony闭幕式")
}
