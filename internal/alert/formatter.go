package alert

import (
	"fmt"
	"strings"
)

func Format(a Alert) string {
	status := formatStatus(a.Status)

	name := a.Labels["alertname"]
	if name == "" {
		name = "Unknown Alert"
	}

	summary := a.Annotations["summary"]
	if summary == "" {
		summary = a.Annotations["description"]
		if summary == "" {
			summary = "No description provided"
		}
	}

	escapedName := escapeMarkdown(name)
	escapedSummary := escapeMarkdown(summary)

	result := fmt.Sprintf("%s\n**%s**\n%s", status, escapedName, escapedSummary)

	if instance := a.Labels["instance"]; instance != "" {
		result += fmt.Sprintf("\n**Instance:** %s", escapeMarkdown(instance))
	}
	if severity := a.Labels["severity"]; severity != "" {
		result += fmt.Sprintf("\n**Severity:** %s", escapeMarkdown(severity))
	}

	if a.GeneratorURL != "" {
		result += fmt.Sprintf("\n\n[→ Открыть в Grafana](%s)", a.GeneratorURL)
	}

	if a.Status == "resolved" && !a.EndsAt.IsZero() {
		result += fmt.Sprintf("\n\n✅ Решено: %s", a.EndsAt.Format("2006-01-02 15:04:05"))
	}

	return result
}

func formatStatus(status string) string {
	switch status {
	case "firing":
		return "🔥 **FIRING**"
	case "resolved":
		return "✅ **RESOLVED**"
	default:
		return strings.ToUpper(status)
	}
}

func escapeMarkdown(text string) string {
	replacer := strings.NewReplacer(
		"_", "\\_", "*", "\\*", "[", "\\[", "]", "\\]", "`", "\\`",
		"\\", "\\\\", "(", "\\(", ")", "\\)", "~", "\\~", ">", "\\>",
		"#", "\\#", "+", "\\+", "-", "\\-", "=", "\\=", "|", "\\|",
		"{", "\\{", "}", "\\}", ".", "\\.", "!", "\\!",
	)
	return replacer.Replace(text)
}
