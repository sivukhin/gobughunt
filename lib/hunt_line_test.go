package lib

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractHuntLines(t *testing.T) {
	huntLines, skipped := ExtractHuntLines([]string{
		`2024/02/06 01:46:35 it seems like your code vanished from compiled binary: func=[validateUserGroups], file=[/repo/internal/dataprovider/dataprovider.go], lines=[2704-2704], snippet:`,
		`return util.NewValidationError(fmt.Sprintf("invalid group type: %v", g.Type))`,
		`2024/02/06 01:46:35 it seems like your code vanished from compiled binary: func=[handleSIGHUP], file=[/repo/internal/service/signals_unix.go], lines=[77-77], snippet:`,
		`logger.Warn(logSender, "", "error reloading common configs: %v", err)`,
	})
	require.Equal(t, []HuntLine{{
		File:      "/repo/internal/dataprovider/dataprovider.go",
		Function:  "validateUserGroups",
		StartLine: 2704,
		EndLine:   2704,
	}, {
		File:      "/repo/internal/service/signals_unix.go",
		Function:  "handleSIGHUP",
		StartLine: 77,
		EndLine:   77,
	}}, huntLines)
	t.Log(huntLines, skipped)
}
