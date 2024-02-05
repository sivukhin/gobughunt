package lib

import (
	"regexp"
	"strconv"
)

type HuntLine struct {
	File      string
	Function  string
	StartLine int
	EndLine   int
}

var (
	huntLineFileRe  = regexp.MustCompile("file=\\[([^]]+)]")
	huntLineRangeRe = regexp.MustCompile("lines=\\[(\\d+)-(\\d+)]")
	huntLineFuncRe  = regexp.MustCompile("func=\\[([^]]+)]")
)

func ExtractHuntLines(rawLines []string) (huntLines []HuntLine, skipped bool) {
	for _, line := range rawLines {
		fileMatch := huntLineFileRe.FindStringSubmatch(line)
		rangeMatch := huntLineRangeRe.FindStringSubmatch(line)
		if len(fileMatch) != 2 || len(rangeMatch) != 3 {
			continue
		}
		startLine, err := strconv.Atoi(rangeMatch[1])
		if err != nil {
			continue
		}
		endLine, err := strconv.Atoi(rangeMatch[2])
		if err != nil {
			continue
		}
		funcMatch := huntLineFuncRe.FindStringSubmatch(line)
		funcName := ""
		if len(funcMatch) == 2 {
			funcName = funcMatch[1]
		}
		huntLines = append(huntLines, HuntLine{
			File:      fileMatch[1],
			StartLine: startLine,
			EndLine:   endLine,
			Function:  funcName,
		})
	}
	return huntLines, false
}
