package replaying

import (
	"github.com/v2pro/koala/recording"
	"github.com/v2pro/koala/countlog"
	"context"
	"bytes"
	"fmt"
)

func (replayingSession *ReplayingSession) MatchOutboundTalk(
	ctx context.Context, lastMatchedIndex int, request []byte) (int, float64, *recording.CallOutbound) {
	unit := 16
	chunks := cutToChunks(request, unit)
	keys := replayingSession.loadKeys()
	scores := make([]int, len(replayingSession.CallOutbounds))
	maxScore := 0
	maxScoreIndex := 0
	for chunkIndex, chunk := range chunks {
		for j, key := range keys {
			if j <= lastMatchedIndex {
				continue
			}
			if len(key) < len(chunk) {
				if scores[j] > 0 {
					scores[j]--
				}
				continue
			}
			pos := bytes.Index(key, chunk)
			if pos >= 0 {
				keys[j] = key[pos:]
				if chunkIndex == 0 && lastMatchedIndex == -1 {
					scores[j] += len(chunks) // first chunk has more weight
				} else {
					scores[j]++
				}
				hasBetterScore := scores[j] > maxScore
				if hasBetterScore {
					maxScore = scores[j]
					maxScoreIndex = j
				}
			} else {
				if scores[j] > 0 {
					scores[j]--
				}
			}
		}
	}
	countlog.Trace("event!replaying.talks_scored",
		"ctx", ctx,
		"lastMatchedIndex", lastMatchedIndex,
		"totalScore", len(chunks),
		"scores", func() interface{} {
			return fmt.Sprintf("%v", scores)
		})
	if maxScore == 0 {
		return -1, 0, nil
	}
	mark := float64(maxScore) / float64(len(chunks))
	if lastMatchedIndex != -1 {
		// not starting from beginning, should have minimal score
		if mark < 0.85 {
			return -1, 0, nil
		}
	} else {
		if mark < 0.1 {
			return -1, 0, nil
		}
	}
	return maxScoreIndex, mark, replayingSession.CallOutbounds[maxScoreIndex]

}

func (replayingSession *ReplayingSession) loadKeys() [][]byte {
	keys := make([][]byte, len(replayingSession.CallOutbounds))
	for i, entry := range replayingSession.CallOutbounds {
		keys[i] = entry.Request
	}
	return keys
}

func cutToChunks(key []byte, unit int) [][]byte {
	chunks := [][]byte{}
	if len(key) > 256 {
		offset := 0
		for {
			strikeStart, strikeLen := findReadableChunk(key[offset:])
			if strikeStart == -1 {
				break
			}
			if strikeLen > 8 {
				firstChunkLen := strikeLen
				if firstChunkLen > 16 {
					firstChunkLen = 16
				}
				chunks = append(chunks, key[offset+strikeStart:offset+strikeStart+firstChunkLen])
				key = key[offset+strikeStart+firstChunkLen:]
				break
			}
			offset += strikeStart + strikeLen
		}
	}
	chunkCount := len(key) / unit
	for i := 0; i < len(key)/unit; i++ {
		chunks = append(chunks, key[i*unit:(i+1)*unit])
	}
	lastChunk := key[chunkCount*unit:]
	if len(lastChunk) > 0 {
		chunks = append(chunks, lastChunk)
	}
	return chunks
}

func findReadableChunk(key []byte) (int, int) {
	start := bytes.IndexFunc(key, func(r rune) bool {
		return r > 31 && r < 127
	})
	if start == -1 {
		return -1, -1
	}
	end := bytes.IndexFunc(key[start:], func(r rune) bool {
		return r <= 31 || r >= 127
	})
	if end == -1 {
		return start, len(key) - start
	}
	return start, end
}