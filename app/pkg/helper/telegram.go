package helper

import (
	"fmt"
	"strings"
)

const specialChars = "_*[]()~`>#+-=|{}.!"

type Context int

const (
	ContextText Context = iota
	ContextInlineCode
	ContextCodeBlock
	ContextLinkText
	ContextLinkURL
)

type Char struct {
	Value string
	Pos   int
}

func EscapeText(text string) string {
	replacer := strings.NewReplacer(
		"_", "\\_", "*", "\\*", "[", "\\[", "]", "\\]",
		"(", "\\(", ")", "\\)", "~", "\\~", "`", "\\`",
		">", "\\>", "#", "\\#", "+", "\\+", "-", "\\-",
		"=", "\\=", "|", "\\|", "{", "\\{", "}", "\\}",
		".", "\\.", "!", "\\!",
	)
	return replacer.Replace(text)
}

func SplitMessage(text string, minChunkSize int) []string {
	var chunks []string
	remaining := text

	for len(remaining) > minChunkSize {
		searchRegion := remaining[minChunkSize:]

		splitIndex := strings.Index(searchRegion, "\n\n")

		if splitIndex != -1 {
			cutPoint := minChunkSize + splitIndex + 2
			chunks = append(chunks, strings.TrimSpace(remaining[:cutPoint]))
			remaining = remaining[cutPoint:]
			continue
		}

		splitIndex = strings.Index(searchRegion, "\n")

		if splitIndex != -1 {
			cutPoint := minChunkSize + splitIndex + 1
			chunks = append(chunks, strings.TrimSpace(remaining[:cutPoint]))
			remaining = remaining[cutPoint:]
			continue
		}

		break
	}

	if len(strings.TrimSpace(remaining)) > 0 {
		chunks = append(chunks, strings.TrimSpace(remaining))
	}

	return chunks
}

func ValidateMarkdownV2(input string) error {
	var (
		stack []Char
		ctx   = ContextText
		runes = []rune(input)
		n     = len(runes)
		i     = 0
	)

	for i < n {
		r := runes[i]

		// 1. Handle Backslash Escaping
		if r == '\\' {
			if i+1 >= n {
				return fmt.Errorf("syntax error at pos %d: trailing backslash", i)
			}
			// Escaped character is valid in any context; skip next character
			i += 2
			continue
		}

		// 2. State-Specific Parsing Logic
		switch ctx {

		case ContextCodeBlock:
			// Check for closing triple backticks ```
			if r == '`' && i+2 < n && runes[i+1] == '`' && runes[i+2] == '`' {
				ctx = ContextText
				i += 3
				continue
			}
			// Only ` and \ need escaping inside code blocks
			if r == '`' {
				return fmt.Errorf("syntax error at pos %d: unescaped backtick inside code block", i)
			}
			i++

		case ContextInlineCode:
			if r == '`' {
				ctx = ContextText
				i++
				continue
			}
			i++

		case ContextLinkURL:
			if r == ')' {
				ctx = ContextText
				i++
				continue
			}
			i++

		case ContextLinkText, ContextText:
			// Transition: Code block (```)
			if r == '`' && i+2 < n && runes[i+1] == '`' && runes[i+2] == '`' {
				// Optional: language specification string can follow here until \n
				ctx = ContextCodeBlock
				i += 3
				continue
			}

			// Transition: Inline code (`)
			if r == '`' {
				ctx = ContextInlineCode
				i++
				continue
			}

			// Transition: Inline Link Text [...]
			if r == '[' {
				stack = append(stack, Char{Value: "[", Pos: i})
				ctx = ContextLinkText
				i++
				continue
			}

			if r == ']' && ctx == ContextLinkText {
				if len(stack) == 0 || stack[len(stack)-1].Value != "[" {
					return fmt.Errorf("syntax error at pos %d: unexpected ']' character", i)
				}

				stack = stack[:len(stack)-1] // Pop '['

				// Link text must be followed immediately by '(' for URL
				if i+1 < n && runes[i+1] == '(' {
					ctx = ContextLinkURL
					i += 2 // Consume ']' and '('
					continue
				}
				return fmt.Errorf("syntax error at pos %d: link bracket ']' not followed by '('", i)
			}

			// Multi-char formatting delimiters: || (Spoiler) and __ (Underline)
			if (r == '|' && i+1 < n && runes[i+1] == '|') || (r == '_' && i+1 < n && runes[i+1] == '_') {
				delim := string(runes[i : i+2])
				if len(stack) > 0 && stack[len(stack)-1].Value == delim {
					stack = stack[:len(stack)-1] // Close tag
				} else {
					stack = append(stack, Char{Value: delim, Pos: i}) // Open tag
				}
				i += 2
				continue
			}

			// Single-char formatting delimiters: *, _, ~, >
			if r == '*' || r == '_' || r == '~' || r == '>' {
				delim := string(r)
				if len(stack) > 0 && stack[len(stack)-1].Value == delim {
					stack = stack[:len(stack)-1] // Close tag
				} else {
					stack = append(stack, Char{Value: delim, Pos: i}) // Open tag
				}
				i++
				continue
			}

			// Reserved Special Characters Validation
			if strings.ContainsRune(specialChars, r) {
				return fmt.Errorf("syntax error at pos %d: reserved character '%c' must be escaped", i, r)
			}

			i++
		}
	}

	// 3. Final State and Stack Cleanliness Verification
	if ctx != ContextText {
		return fmt.Errorf("syntax error: unclosed block context %v", ctx)
	}
	if len(stack) > 0 {
		char := stack[len(stack)-1]
		return fmt.Errorf("syntax error: unclosed formatting tag '%s', in: %s", char.Value, string(runes[char.Pos-10:char.Pos+10]))
	}

	return nil
}
