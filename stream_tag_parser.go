package main

import (
	"context"
	"strings"
)

const (
	DefaultStartTagPrefix = "<tag="
	DefaultStartTagSuffix = ">"
	DefaultEndTag         = "</tag>"
)

var DefaultTagConfig = TagConfig{
	StartTagPrefix:    []rune(DefaultStartTagPrefix),
	StartTagSuffix:    []rune(DefaultStartTagSuffix),
	EndTag:            []rune(DefaultEndTag),
	HasEnd:            true,
	StartTagPrefixStr: DefaultStartTagPrefix,
	StartTagSuffixStr: DefaultStartTagSuffix,
	EndTagStr:         DefaultEndTag,
}

type TagConfig struct {
	// tags
	StartTagPrefix    []rune `json:"start_tag_prefix,omitempty"` // 如 "<cot" 或 "("
	StartTagSuffix    []rune `json:"start_tag_suffix,omitempty"` // 如 ">" 或 ""
	EndTag            []rune `json:"end_tag,omitempty"`          // 如 "</cot>" 或 ")"
	HasEnd            bool   `json:"has_end,omitempty"`          // 是否有结束标签
	StartTagPrefixStr string `json:"start_tag_prefix_str,omitempty"`
	StartTagSuffixStr string `json:"start_tag_suffix_str,omitempty"`
	EndTagStr         string `json:"end_tag_str,omitempty"`
}

type TagText struct {
	NormalText    string
	TagText       string
	TagTextBuffer string // 缓存的不完整的tag文本
	TagEnd        bool
	NewTag        bool
}

// StreamTagParser structure
type StreamTagParser struct {
	AppKey    string
	tagConfig TagConfig

	// state machine
	tryStart bool // try to match start tag prefix
	inTag    bool // parsing tag content
	tryEnd   bool // try to match end tag suffix

	startBuf strings.Builder // store start tag prefix match
	tagBuf   strings.Builder // store tag content
	tagText  string
	endBuf   strings.Builder // store end tag suffix match
}

func NewStreamTagParser(ctx context.Context, appkey string) *StreamTagParser {

	return &StreamTagParser{
		AppKey:    appkey,
		tagConfig: DefaultTagConfig,
	}

}

// parse input text to split plain text and tag text
func (p *StreamTagParser) Parse(ctx context.Context, chunk string) (tagTexts []TagText) {
	// empty chunk, no need to parse
	if chunk == "" {
		return []TagText{
			{
				NormalText: chunk,
				TagText:    "",
				TagEnd:     false,
				NewTag:     false,
			},
		}
	}

	var textBuf strings.Builder // store normal text
	// parse chunk by rune
	runes := []rune(chunk)
	for i := 0; i < len(runes); i++ {
		ch := runes[i]

		// Status 1: normal text
		if !p.inTag && !p.tryStart && ch == p.tagConfig.StartTagPrefix[0] {
			// in case start tag prefix is single char, we can match it directly
			if len(p.tagConfig.StartTagPrefix) == 1 {
				p.inTag = true
				p.tryStart = false
				p.tagBuf.Reset()
				p.tagBuf.WriteRune(ch)
				continue
			}
			p.tryStart = true
			p.startBuf.Reset()
			p.startBuf.WriteRune(ch)
			continue
		}

		if p.tryStart {
			// continue match → continue waiting for characters
			if p.handleTryStart(ch, &textBuf) {
				// complete match → enter tag content parsing mode
				continue
			}
			// otherwise, match prefix not found → backtrack output
			continue
		}

		// Status 2: parse tag content
		if p.inTag {
			newTagText, ok := p.handleInTag(ch)
			// matched tag text
			if ok {
				tagTexts = append(tagTexts, TagText{
					NormalText:    textBuf.String(),
					TagText:       p.tagText,
					TagTextBuffer: p.tagBuf.String(),
					TagEnd:        false,
					NewTag:        true,
				})
				textBuf.Reset()

				p.inTag = false
				p.tagBuf.Reset()
				p.tagText = newTagText
			}
			continue
		}

		// Status 3: check end tag
		if p.tagConfig.HasEnd && p.tryEnd {
			if p.handleTryEnd(ctx, ch, &textBuf) {
				tagTexts = append(tagTexts, TagText{
					NormalText:    textBuf.String(),
					TagText:       p.tagText,
					TagTextBuffer: p.tagBuf.String(),
					TagEnd:        true,
					NewTag:        false,
				})
				textBuf.Reset()

				// when end tag matched, reset tagBuf
				p.tagBuf.Reset()
				p.tagText = ""
				// when end tag matched, flush segment text
				continue
			}
			continue
		}

		// Status 4: normal text
		textBuf.WriteRune(ch)
	}
	// when TagText or TagTextBuffer not empty, need return
	tagTexts = append(tagTexts, TagText{
		NormalText:    textBuf.String(),
		TagText:       p.tagText,
		TagTextBuffer: p.tagBuf.String(),
		TagEnd:        false,
		NewTag:        p.inTag,
	})

	return tagTexts
}

// handleTryStart handle StartTag prefix match logic
func (p *StreamTagParser) handleTryStart(ch rune, textBuf *strings.Builder) bool {
	p.startBuf.WriteRune(ch)
	prefix := p.startBuf.String()

	// 1. match prefix not found → backtrack output normal text
	if !strings.HasPrefix(p.tagConfig.StartTagPrefixStr, prefix) {
		p.tryStart = false
		p.startBuf.Reset()
		if strings.HasPrefix(p.tagConfig.EndTagStr, prefix) {
			p.tryEnd = true
			p.endBuf.WriteString(prefix)
			return false
		}
		// backtrack output normal text
		textBuf.WriteString(prefix)
		return false
	}

	// 2. complete match → enter tag content parsing mode
	if prefix == p.tagConfig.StartTagPrefixStr {
		p.inTag = true
		p.tryStart = false
		p.tagBuf.Reset()
		p.tagBuf.WriteString(prefix)
		return true
	}

	// 3. continue match → continue waiting for characters
	return false
}

// handleInTag handle tag content parsing logic
func (p *StreamTagParser) handleInTag(ch rune) (tagText string, ok bool) {
	p.tagBuf.WriteRune(ch)
	tagStr := p.tagBuf.String()

	// no end tag, match suffix → enter tag content parsing mode, example: ()
	// todo: expand to support multiple chars suffix
	if ch == p.tagConfig.StartTagSuffix[0] {
		tagStr = strings.TrimPrefix(tagStr, p.tagConfig.StartTagPrefixStr)
		tagText = strings.TrimSuffix(tagStr, p.tagConfig.StartTagSuffixStr)
		tagText = strings.TrimSpace(tagText)
		p.inTag = false
		p.tagBuf.Reset()
		return tagText, true
	}
	return "", false
}

// handleTryEnd handle EndTag suffix match logic
func (p *StreamTagParser) handleTryEnd(ctx context.Context, ch rune, textBuf *strings.Builder) (ok bool) {
	p.endBuf.WriteRune(ch)
	endTag := p.endBuf.String()
	// check if the end tag prefix matched
	if !strings.HasPrefix(p.tagConfig.EndTagStr, endTag) {
		p.tryEnd = false
		p.endBuf.Reset()
		textBuf.WriteString(endTag)
		return false
	}

	// end tag matched
	if endTag == p.tagConfig.EndTagStr {
		p.tryEnd = false
		p.endBuf.Reset()
		return true
	}

	return false
}
