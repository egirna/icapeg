package http_server

import (
	"testing"

	utils "icapeg/consts"
	"icapeg/logging"
	services_utilities "icapeg/service/services-utilities"

	"go.uber.org/zap"
)

// equal checks if two slices of strings are equal
func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// init initializes the logger to avoid nil pointer dereference during tests
func init() {
	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	logging.Logger = logger
}

// TestInitExtsArr tests the InitExtsArr function
func TestInitExtsArr(t *testing.T) {
	tests := []struct {
		name          string
		processExts   []string
		rejectExts    []string
		bypassExts    []string
		expectedOrder []services_utilities.Extension
	}{
		{
			name:        "Normal case with different extensions",
			processExts: []string{"pdf", "zip", "com"},
			rejectExts:  []string{"docx"},
			bypassExts:  []string{"*"},
			expectedOrder: []services_utilities.Extension{
				{Name: utils.RejectExts, Exts: []string{"docx"}},
				{Name: utils.ProcessExts, Exts: []string{"pdf", "zip", "com"}},
				{Name: utils.BypassExts, Exts: []string{"*"}},
			},
		},
		{
			name:        "Case with reject having only asterisk",
			processExts: []string{"pdf", "zip", "com"},
			rejectExts:  []string{"*"},
			bypassExts:  []string{"docx"},
			expectedOrder: []services_utilities.Extension{
				{Name: utils.ProcessExts, Exts: []string{"pdf", "zip", "com"}},
				{Name: utils.BypassExts, Exts: []string{"docx"}},
				{Name: utils.RejectExts, Exts: []string{"*"}},
			},
		},
		{
			name:        "Case with all having multiple extensions",
			processExts: []string{"pdf", "zip"},
			rejectExts:  []string{"docx", "xlsx"},
			bypassExts:  []string{"exe", "bin"},
			expectedOrder: []services_utilities.Extension{
				{Name: utils.ProcessExts, Exts: []string{"pdf", "zip"}},
				{Name: utils.RejectExts, Exts: []string{"docx", "xlsx"}},
				{Name: utils.BypassExts, Exts: []string{"exe", "bin"}},
			},
		},
		{
			name:        "Edge case with empty arrays",
			processExts: []string{},
			rejectExts:  []string{"docx"},
			bypassExts:  []string{"*"},
			expectedOrder: []services_utilities.Extension{
				{Name: utils.RejectExts, Exts: []string{"docx"}},
				{Name: utils.ProcessExts, Exts: []string{}},
				{Name: utils.BypassExts, Exts: []string{"*"}},
			},
		},
		{
			name:        "Case with all categories having only asterisk",
			processExts: []string{"*"},
			rejectExts:  []string{"*"},
			bypassExts:  []string{"*"},
			expectedOrder: []services_utilities.Extension{
				{Name: utils.ProcessExts, Exts: []string{"*"}},
				{Name: utils.RejectExts, Exts: []string{"*"}},
				{Name: utils.BypassExts, Exts: []string{"*"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := services_utilities.InitExtsArr(tt.processExts, tt.rejectExts, tt.bypassExts)

			// Check the last element separately
			if result[len(result)-1].Name != tt.expectedOrder[len(tt.expectedOrder)-1].Name || !equal(result[len(result)-1].Exts, tt.expectedOrder[len(tt.expectedOrder)-1].Exts) {
				t.Errorf("Test case '%s' failed: expected last element %v, got %v", tt.name, tt.expectedOrder[len(tt.expectedOrder)-1], result[len(result)-1])
			}

			// Check the first two elements ignoring the order
			if !(result[0].Name == tt.expectedOrder[0].Name && equal(result[0].Exts, tt.expectedOrder[0].Exts) &&
				result[1].Name == tt.expectedOrder[1].Name && equal(result[1].Exts, tt.expectedOrder[1].Exts)) &&
				!(result[0].Name == tt.expectedOrder[1].Name && equal(result[0].Exts, tt.expectedOrder[1].Exts) &&
					result[1].Name == tt.expectedOrder[0].Name && equal(result[1].Exts, tt.expectedOrder[0].Exts)) {
				t.Errorf("Test case '%s' failed: expected first two elements %v and %v, got %v and %v", tt.name, tt.expectedOrder[0], tt.expectedOrder[1], result[0], result[1])
			}
		})
	}
}
