package service

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
)

// The GitHub call-to-action in report footers is the only path from a generated
// report back to the repository. Most users install via `uvx pyscn` and never
// visit GitHub, so silently dropping these links costs the project its main
// discovery route. These tests exist to make such a regression loud.

func TestRepositoryURLPointsAtTheRepo(t *testing.T) {
	const want = "https://github.com/ludo-technologies/pyscn"
	if RepositoryURL != want {
		t.Errorf("RepositoryURL = %q, want %q", RepositoryURL, want)
	}
}

func TestGenerateStarLinkIsAnAnchorToTheRepo(t *testing.T) {
	link := GenerateStarLink()

	if !strings.Contains(link, `href="`+RepositoryURL+`"`) {
		t.Errorf("star link does not href the repository: %q", link)
	}
	if !strings.Contains(link, `rel="noopener noreferrer"`) {
		t.Errorf("star link opens a new tab without noopener/noreferrer: %q", link)
	}
	// A remote badge image would make reports phone home on every open.
	if strings.Contains(link, "<img") || strings.Contains(link, "shields.io") {
		t.Errorf("star link should be a plain anchor, not a remote image: %q", link)
	}
}

func TestGenerateHTMLFooterIncludesStarLink(t *testing.T) {
	footer := GenerateHTMLFooter()

	if !strings.Contains(footer, RepositoryURL) {
		t.Errorf("standard HTML footer lost its repository link:\n%s", footer)
	}
}

// The analyze report is the one every `pyscn analyze` user sees, so it matters
// most. It builds its own markup instead of calling GenerateHTMLFooter.
func TestAnalyzeHTMLTemplateIncludesStarLink(t *testing.T) {
	var buf bytes.Buffer
	if err := writeAnalyzeHTML(&domain.AnalyzeResponse{}, &buf); err != nil {
		t.Fatalf("render analyze report: %v", err)
	}
	if !strings.Contains(buf.String(), RepositoryURL) {
		t.Error("analyze HTML report has no link back to the repository")
	}
	if !strings.Contains(buf.String(), "<footer>") {
		t.Error("analyze HTML report is missing its footer block")
	}
}
