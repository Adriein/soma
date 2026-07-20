package helper

import "strings"

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
