package chunker

import (
	"strings"
)

const overlapRatio = 0.15

func ChunkText(text string, chunkSize int) []string {
	var chunks []string
	var current strings.Builder
	var currentWords []string

	words := strings.FieldsSeq(text)

	for word := range words {
		// Plus 1 accounts for the space
		if current.Len()+len(word)+1 > chunkSize {
			// Finalize current chunk
			chunks = append(chunks, current.String())

			// Calculate roughly how many characters equate to 15% of the chunk
			overlapSize := int(float64(current.Len()) * overlapRatio)

			// Track character count backwards to include enough words for overlap
			charCount := 0
			var overlapWords []string
			for i := len(currentWords) - 1; i >= 0; i-- {
				wordLen := len(currentWords[i])
				charCount += wordLen
				if i != len(currentWords)-1 {
					charCount++ // Add space character
				}
				if charCount > overlapSize {
					break
				}
				overlapWords = append([]string{currentWords[i]}, overlapWords...)
			}

			// Start new chunk with overlapping words
			current.Reset()
			currentWords = overlapWords
			for i, ow := range overlapWords {
				if i > 0 {
					current.WriteByte(' ')
				}
				current.WriteString(ow)
			}
		}

		if current.Len() > 0 {
			current.WriteByte(' ')
		}
		current.WriteString(word)
		currentWords = append(currentWords, word)
	}

	if current.Len() > 0 {
		chunks = append(chunks, current.String())
	}

	return chunks
}
