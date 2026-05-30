package service

import (
	"strings"
	"testing"
)

func TestGenerateMetricCardEscapesValueAndLabel(t *testing.T) {
	html := GenerateMetricCard(`<script>alert("x")</script>`, `<b>Total</b>`)

	if strings.Contains(html, "<script>") || strings.Contains(html, "<b>Total</b>") {
		t.Fatalf("expected raw metric content to be escaped, got %s", html)
	}
	if !strings.Contains(html, `&lt;script&gt;alert(&#34;x&#34;)&lt;/script&gt;`) {
		t.Fatalf("expected escaped metric value, got %s", html)
	}
	if !strings.Contains(html, `&lt;b&gt;Total&lt;/b&gt;`) {
		t.Fatalf("expected escaped metric label, got %s", html)
	}
}

func TestGenerateMetricCardHTMLPreservesTrustedValue(t *testing.T) {
	html := generateMetricCardHTML(trustedHTML(`<span class="status-badge">ok</span>`), `<b>Status</b>`)

	if !strings.Contains(html, `<span class="status-badge">ok</span>`) {
		t.Fatalf("expected trusted metric HTML value to be preserved, got %s", html)
	}
	if strings.Contains(html, "<b>Status</b>") || !strings.Contains(html, `&lt;b&gt;Status&lt;/b&gt;`) {
		t.Fatalf("expected metric label to be escaped, got %s", html)
	}
}

func TestGenerateTabButtonEscapesLabelAndSanitizesID(t *testing.T) {
	html := GenerateTabButton(`sys' onclick='alert(1)`, `<b>System</b>`, true)

	if strings.Contains(html, `sys' onclick='alert(1)`) {
		t.Fatalf("expected tab id to be sanitized, got %s", html)
	}
	if !strings.Contains(html, `showTab('sys__onclick__alert_1_', this)`) {
		t.Fatalf("expected sanitized tab id in onclick, got %s", html)
	}
	if strings.Contains(html, "<b>System</b>") || !strings.Contains(html, `&lt;b&gt;System&lt;/b&gt;`) {
		t.Fatalf("expected tab label to be escaped, got %s", html)
	}
}
