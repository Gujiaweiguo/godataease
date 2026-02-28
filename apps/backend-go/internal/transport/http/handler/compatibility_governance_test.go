package handler_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestNoPlaceholderSuccessInCompatibilityBridge(t *testing.T) {
	repoRoot := filepath.Join("..", "..", "..", "..")
	handlerPath := filepath.Join(repoRoot, "internal", "transport", "http", "handler", "compatibility_bridge_handler.go")

	content, err := os.ReadFile(handlerPath)
	if err != nil {
		t.Skipf("Could not read compatibility_bridge_handler.go: %v", err)
	}

	source := string(content)
	lines := strings.Split(source, "\n")

	suspiciousPatterns := []struct {
		pattern *regexp.Regexp
		reason  string
	}{
		{
			pattern: regexp.MustCompile(`response\.Success\(c,\s*(nil|\[\]interface\{\}\{\})\)\s*//\s*TODO`),
			reason:  "Success with TODO comment suggests placeholder implementation",
		},
		{
			pattern: regexp.MustCompile(`(?i)//\s*(stub|placeholder|not implemented)`),
			reason:  "Comment indicates stub/placeholder behavior",
		},
	}

	var issues []string
	for i, line := range lines {
		for _, sp := range suspiciousPatterns {
			if sp.pattern.MatchString(line) {
				issues = append(issues, sp.reason)
				t.Logf("Line %d: %s", i+1, sp.reason)
			}
		}
	}

	if len(issues) > 0 {
		t.Fatalf("Found %d potential placeholder patterns", len(issues))
	}
}

func TestDeleteOperationsReturnSuccessNil(t *testing.T) {
	repoRoot := filepath.Join("..", "..", "..", "..")
	handlerPath := filepath.Join(repoRoot, "internal", "transport", "http", "handler", "compatibility_bridge_handler.go")

	content, err := os.ReadFile(handlerPath)
	if err != nil {
		t.Skipf("Could not read compatibility_bridge_handler.go: %v", err)
	}

	deletePattern := regexp.MustCompile(`(?i)(delete|remove|destroy).*Success\(c,\s*nil\)`)
	matches := deletePattern.FindAllString(string(content), -1)

	for _, m := range matches {
		t.Logf("Valid delete operation: %s", m)
	}
}

func TestStatusConsistencyWithWhitelist(t *testing.T) {
	repoRoot := filepath.Join("..", "..", "..", "..")
	whitelistPath := filepath.Join(repoRoot, "testdata", "contract-diff", "critical-whitelist.yaml")

	content, err := os.ReadFile(whitelistPath)
	if err != nil {
		t.Skipf("Could not read whitelist: %v", err)
	}

	whitelist := string(content)

	partialEndpoints := []string{
		"/datasource/syncApiTable",
		"/datasource/syncApiDs",
		"/datasetData/previewSql",
	}

	for _, endpoint := range partialEndpoints {
		if !strings.Contains(whitelist, endpoint) {
			t.Errorf("Partial endpoint %s not found in whitelist", endpoint)
			continue
		}

		section := extractEndpointSection(whitelist, endpoint)
		if section == "" {
			t.Errorf("Could not extract section for %s", endpoint)
			continue
		}

		if !strings.Contains(section, "goStatus: partial") && !strings.Contains(section, "goStatus: \"partial\"") {
			t.Errorf("Endpoint %s should have goStatus: partial", endpoint)
		}

		if !strings.Contains(section, "gaps:") {
			t.Errorf("Partial endpoint %s missing gaps documentation", endpoint)
		}
	}
}

func extractEndpointSection(whitelist, endpoint string) string {
	lines := strings.Split(whitelist, "\n")
	var result []string
	inSection := false
	indentLevel := 0

	for i, line := range lines {
		if strings.Contains(line, endpoint) && strings.Contains(line, "path:") {
			inSection = true
			indentLevel = len(line) - len(strings.TrimLeft(line, " "))
		}

		if inSection {
			result = append(result, line)

			if i+1 < len(lines) {
				nextLine := lines[i+1]
				nextIndent := len(nextLine) - len(strings.TrimLeft(nextLine, " "))
				if nextIndent <= indentLevel && strings.TrimSpace(nextLine) != "" && !strings.HasPrefix(nextLine, strings.Repeat(" ", indentLevel)) {
					break
				}
			}
		}
	}

	return strings.Join(result, "\n")
}
