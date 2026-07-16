package client

import (
	"fmt"
	"io"
)

// ResultWriter writes testcase results to a writer
func ResultWriter(w io.Writer, results []TestCase) {
	var passMsg = map[bool]string{true: "PASS", false: "FAIL"}
	for _, result := range results {
		fmt.Fprintf(w, "=== %s: %s\n", passMsg[result.Pass], result.Id)
		if !result.Pass {
			for _, msg := range result.Fail {
				fmt.Fprintf(w, "\t %s\n", msg)
			}
		}
	}
}
