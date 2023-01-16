package models

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_sanitizeQuery(t *testing.T) {
	sanitizeTests := []struct {
		title  string
		input  string
		output string
	}{
		{"removes comments, unused whitespace, and ;", `
-- whatever this is
SELECT * from         whatever;
-- another extraneous comment.
`, "SELECT * from whatever"},
		{"option to not lowercase a query", `
-- whatever this is
SELECT * from         whateveR;
-- another extraneous comment.        
        `, "SELECT * from whateveR"},
		{"removes extraneous spaces from columns", `
-- whatever this is
SELECT 1, 2,     3 from         whateveR;
-- another extraneous comment.        
        `, "SELECT 1,2,3 from whateveR"},
	}

	for _, sanitizeTest := range sanitizeTests {
		t.Run(sanitizeTest.title, func(t *testing.T) {
			require.Equal(t, sanitizeTest.output, sanitizeQuery(sanitizeTest.input))
		})
	}
}
