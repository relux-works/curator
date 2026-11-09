package profilegraph

import (
	"bytes"
	"strings"
)

type markdownLine struct {
	start, end int
	text       string
}
type markdownHeading struct {
	start, level int
	text         string
}

func (b *build) markdown(module *Node, raw []byte) error {
	var previous markdownLine
	headings := []markdownHeading{}
	quotes := []Range{}
	fence := byte(0)
	fenceLength := 0
	quoteStart := -1
	previousParagraph := false
	for offset, i := 0, 0; offset < len(raw); i++ {
		end := bytes.IndexByte(raw[offset:], '\n')
		if end < 0 {
			end = len(raw)
		} else {
			end += offset + 1
		}
		l := markdownLine{offset, end, strings.TrimSuffix(strings.TrimSuffix(string(raw[offset:end]), "\n"), "\r")}
		offset = end
		prior := previous
		previous = l
		if len(headings)+len(quotes) >= b.r.Limits.Nodes {
			return refusal("graph_limit_exceeded")
		}

		eligiblePrevious := previousParagraph
		previousParagraph = false
		text := strings.TrimLeft(l.text, " ")
		indent := len(l.text) - len(text)
		if indent <= 3 && len(text) >= 3 && (text[0] == '`' || text[0] == '~') {
			count := 0
			for count < len(text) && text[count] == text[0] {
				count++
			}
			if count >= 3 {
				if fence == 0 {
					fence = text[0]
					fenceLength = count
					if quoteStart >= 0 {
						quotes = append(quotes, Range{quoteStart, l.start})
						quoteStart = -1
					}
					continue
				} else if fence == text[0] && count >= fenceLength && strings.TrimSpace(text[count:]) == "" {
					fence = 0
					continue
				}
			}
		}
		if fence != 0 {
			continue
		}
		quote := indent <= 3 && strings.HasPrefix(text, ">")
		if quote && quoteStart < 0 {
			quoteStart = l.start
		}
		if !quote && quoteStart >= 0 {
			quotes = append(quotes, Range{quoteStart, l.start})
			quoteStart = -1
		}
		if indent > 3 || quote {
			continue
		}
		level := 0
		for level < len(text) && text[level] == '#' {
			level++
		}
		if level >= 1 && level <= 6 && (level == len(text) || text[level] == ' ' || text[level] == '\t') {
			title := strings.TrimSpace(text[level:])
			if end := strings.LastIndex(title, " #"); end >= 0 && strings.Trim(title[end+1:], "#") == "" {
				title = strings.TrimSpace(title[:end])
			}
			headings = append(headings, markdownHeading{l.start, level, title})
		} else if eligiblePrevious && i > 0 && strings.TrimSpace(text) != "" && (strings.Trim(text, "=") == "" || strings.Trim(text, "-") == "") {
			title := strings.TrimSpace(prior.text)
			if title != "" && !strings.HasPrefix(title, ">") && !strings.HasPrefix(title, "#") {
				level := 1
				if text[0] == '-' {
					level = 2
				}
				headings = append(headings, markdownHeading{prior.start, level, title})
			}
		} else {
			previousParagraph = strings.TrimSpace(text) != ""
		}
	}
	if quoteStart >= 0 {
		quotes = append(quotes, Range{quoteStart, len(raw)})
	}
	for i, h := range headings {
		end := len(raw)
		for _, next := range headings[i+1:] {
			if next.level <= h.level {
				end = next.start
				break
			}
		}
		n := b.add("section", h.text, module.Source, module.Layer, module.ID, itoa(h.start), itoa(end))
		n.Heading = &Heading{h.level, h.text}
		n.SourceRanges = []Range{{h.start, end}}
		n.Content = measured(raw[h.start:end])
		b.edge(module.ID, n.ID, "section", i, nil)
	}
	for i, q := range quotes {
		n := b.add("quote-block", "quote", module.Source, module.Layer, module.ID, itoa(q.Start), itoa(q.End))
		n.SourceRanges = []Range{q}
		n.Content = measured(raw[q.Start:q.End])
		b.edge(module.ID, n.ID, "quote", i, nil)
	}
	return b.err
}
