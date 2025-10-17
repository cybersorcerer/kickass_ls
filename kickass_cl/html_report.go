package main

import (
	"fmt"
	"html"
	"os"
	"strings"
	"time"
)

// GenerateHTMLReport generates an HTML report from test results
func GenerateHTMLReport(results []TestResult, outputPath string, suiteName string) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create HTML report: %w", err)
	}
	defer f.Close()

	// Count results
	totalTests := len(results)
	passed := 0
	failed := 0
	errors := 0
	for _, r := range results {
		switch r.Status {
		case "PASS":
			passed++
		case "FAIL":
			failed++
		case "ERROR":
			errors++
		}
	}

	// Generate HTML
	fmt.Fprint(f, htmlHeader(suiteName))
	fmt.Fprint(f, htmlSummary(totalTests, passed, failed, errors))
	fmt.Fprint(f, htmlTestResults(results))
	fmt.Fprint(f, htmlFooter())

	return nil
}

func htmlHeader(suiteName string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Test Report: %s</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            margin: 0;
            padding: 20px;
            background: #f5f5f5;
            color: #333;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        h1 {
            margin: 0 0 10px 0;
            color: #2c3e50;
        }
        .subtitle {
            color: #7f8c8d;
            margin-bottom: 30px;
        }
        .summary {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        .summary-card {
            padding: 20px;
            border-radius: 6px;
            text-align: center;
        }
        .summary-card.total {
            background: #3498db;
            color: white;
        }
        .summary-card.passed {
            background: #2ecc71;
            color: white;
        }
        .summary-card.failed {
            background: #e74c3c;
            color: white;
        }
        .summary-card.error {
            background: #e67e22;
            color: white;
        }
        .summary-card .number {
            font-size: 48px;
            font-weight: bold;
            margin-bottom: 5px;
        }
        .summary-card .label {
            font-size: 14px;
            opacity: 0.9;
        }
        .test-list {
            border: 1px solid #ddd;
            border-radius: 6px;
            overflow: hidden;
        }
        .test-item {
            border-bottom: 1px solid #ddd;
            padding: 15px 20px;
            display: grid;
            grid-template-columns: 40px 1fr 100px 100px;
            gap: 15px;
            align-items: center;
            transition: background 0.2s;
        }
        .test-item:last-child {
            border-bottom: none;
        }
        .test-item:hover {
            background: #f8f9fa;
        }
        .test-item.passed {
            background: #f0fdf4;
        }
        .test-item.failed {
            background: #fef2f2;
        }
        .test-item.error {
            background: #fff7ed;
        }
        .status-icon {
            font-size: 24px;
            text-align: center;
        }
        .test-name {
            font-weight: 500;
            color: #2c3e50;
        }
        .test-description {
            color: #7f8c8d;
            font-size: 14px;
            margin-top: 5px;
        }
        .test-type {
            font-size: 12px;
            padding: 4px 8px;
            border-radius: 4px;
            background: #ecf0f1;
            color: #34495e;
            text-align: center;
        }
        .test-duration {
            color: #95a5a6;
            font-size: 14px;
            text-align: right;
        }
        .test-details {
            grid-column: 2 / -1;
            margin-top: 10px;
            padding: 15px;
            background: #f8f9fa;
            border-radius: 4px;
            font-size: 14px;
        }
        .test-message {
            color: #c0392b;
            font-family: monospace;
            white-space: pre-wrap;
        }
        .timestamp {
            text-align: center;
            color: #95a5a6;
            margin-top: 30px;
            font-size: 14px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Test Report: %s</h1>
        <div class="subtitle">Generated on %s</div>
`, html.EscapeString(suiteName), html.EscapeString(suiteName), time.Now().Format("2006-01-02 15:04:05"))
}

func htmlSummary(total, passed, failed, errors int) string {
	return fmt.Sprintf(`
        <div class="summary">
            <div class="summary-card total">
                <div class="number">%d</div>
                <div class="label">Total Tests</div>
            </div>
            <div class="summary-card passed">
                <div class="number">%d</div>
                <div class="label">Passed</div>
            </div>
            <div class="summary-card failed">
                <div class="number">%d</div>
                <div class="label">Failed</div>
            </div>
            <div class="summary-card error">
                <div class="number">%d</div>
                <div class="label">Errors</div>
            </div>
        </div>
`, total, passed, failed, errors)
}

func htmlTestResults(results []TestResult) string {
	var sb strings.Builder
	sb.WriteString(`        <div class="test-list">`)
	sb.WriteString("\n")

	for _, result := range results {
		statusClass := strings.ToLower(result.Status)
		statusIcon := "❓"
		switch result.Status {
		case "PASS":
			statusIcon = "✅"
		case "FAIL":
			statusIcon = "❌"
		case "ERROR":
			statusIcon = "⚠️"
		}

		sb.WriteString(fmt.Sprintf(`            <div class="test-item %s">`, statusClass))
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf(`                <div class="status-icon">%s</div>`, statusIcon))
		sb.WriteString("\n")
		sb.WriteString(`                <div>`)
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf(`                    <div class="test-name">%s</div>`, html.EscapeString(result.TestCase.Name)))
		sb.WriteString("\n")
		if result.TestCase.Description != "" {
			sb.WriteString(fmt.Sprintf(`                    <div class="test-description">%s</div>`, html.EscapeString(result.TestCase.Description)))
			sb.WriteString("\n")
		}
		if result.Message != "" && result.Status != "PASS" {
			sb.WriteString(`                    <div class="test-details">`)
			sb.WriteString("\n")
			sb.WriteString(fmt.Sprintf(`                        <div class="test-message">%s</div>`, html.EscapeString(result.Message)))
			sb.WriteString("\n")
			sb.WriteString(`                    </div>`)
			sb.WriteString("\n")
		}
		sb.WriteString(`                </div>`)
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf(`                <div class="test-type">%s</div>`, html.EscapeString(result.TestCase.Type)))
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf(`                <div class="test-duration">%v</div>`, result.Duration.Round(time.Millisecond)))
		sb.WriteString("\n")
		sb.WriteString(`            </div>`)
		sb.WriteString("\n")
	}

	sb.WriteString(`        </div>`)
	sb.WriteString("\n")
	return sb.String()
}

func htmlFooter() string {
	return `
        <div class="timestamp">
            Report generated by kickass_cl LSP Test Client
        </div>
    </div>
</body>
</html>
`
}
