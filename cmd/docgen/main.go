package main

import (
	"bytes"
	"fmt"
	"html"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type TOCItem struct {
	ID    string
	Text  string
	Level int // 2 or 3
}

func main() {
	mdDir := "docs/md"
	htmlDir := "docs/html"

	if err := os.MkdirAll(htmlDir, 0755); err != nil {
		log.Fatalf("Failed to create html output dir: %v", err)
	}

	count := 0
	err := filepath.WalkDir(mdDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}

		relPath, err := filepath.Rel(mdDir, path)
		if err != nil {
			return err
		}

		outRelPath := strings.TrimSuffix(relPath, ".md") + ".html"
		outPath := filepath.Join(htmlDir, outRelPath)

		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return err
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		isIndex := relPath == "index.md"
		htmlDoc := renderMarkdownToHTML(string(content), relPath, isIndex)

		if err := os.WriteFile(outPath, []byte(htmlDoc), 0644); err != nil {
			return err
		}

		log.Printf("Generated: %s -> %s", path, outPath)
		count++
		return nil
	})

	if err != nil {
		log.Fatalf("Error generating docs: %v", err)
	}

	log.Printf("Successfully generated %d HTML documentation pages.", count)
}

func renderMarkdownToHTML(md string, relPath string, isIndex bool) string {
	lines := strings.Split(md, "\n")
	var bodyBuf bytes.Buffer
	var toc []TOCItem

	inCodeBlock := false
	codeBlockLang := ""
	var codeBuf bytes.Buffer

	inTable := false
	var tableLines []string

	inList := false
	listType := "" // "ul" or "ol"

	inBlockquote := false

	pageTitle := "Go DDD Documentation"

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Code Block Handling
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			if inCodeBlock {
				// End code block
				inCodeBlock = false
				bodyBuf.WriteString(fmt.Sprintf("<pre><code class=\"language-%s\">%s</code></pre>\n",
					html.EscapeString(codeBlockLang),
					html.EscapeString(codeBuf.String())))
				codeBuf.Reset()
				codeBlockLang = ""
				continue
			} else {
				// Close open elements if any
				closeList(&bodyBuf, &inList, &listType)
				closeTable(&bodyBuf, &inTable, &tableLines)
				closeBlockquote(&bodyBuf, &inBlockquote)

				inCodeBlock = true
				codeBlockLang = strings.TrimPrefix(strings.TrimSpace(line), "```")
				continue
			}
		}

		if inCodeBlock {
			codeBuf.WriteString(line + "\n")
			continue
		}

		// Comment markers like <!-- metadata ... --> or <!-- covers: ... -->
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "<!--") && strings.HasSuffix(trimmed, "-->") {
			closeList(&bodyBuf, &inList, &listType)
			closeTable(&bodyBuf, &inTable, &tableLines)
			closeBlockquote(&bodyBuf, &inBlockquote)
			bodyBuf.WriteString(fmt.Sprintf("<div class=\"doc-comment\">%s</div>\n", html.EscapeString(line)))
			continue
		}

		// Table handling
		if strings.HasPrefix(trimmed, "|") && strings.HasSuffix(trimmed, "|") {
			closeList(&bodyBuf, &inList, &listType)
			closeBlockquote(&bodyBuf, &inBlockquote)
			inTable = true
			tableLines = append(tableLines, line)
			continue
		} else if inTable {
			closeTable(&bodyBuf, &inTable, &tableLines)
		}

		// Blockquote handling
		if strings.HasPrefix(trimmed, ">") {
			closeList(&bodyBuf, &inList, &listType)
			if !inBlockquote {
				inBlockquote = true
				bodyBuf.WriteString("blockquote>\n")
			}
			quoteText := strings.TrimSpace(strings.TrimPrefix(trimmed, ">"))
			bodyBuf.WriteString(fmt.Sprintf("<p>%s</p>\n", parseInline(quoteText)))
			continue
		} else if inBlockquote {
			closeBlockquote(&bodyBuf, &inBlockquote)
		}

		// Headings
		if strings.HasPrefix(line, "# ") {
			closeList(&bodyBuf, &inList, &listType)
			pageTitle = strings.TrimPrefix(line, "# ")
			bodyBuf.WriteString(fmt.Sprintf("<h1>%s</h1>\n", parseInline(pageTitle)))
			continue
		}

		if strings.HasPrefix(line, "## ") {
			closeList(&bodyBuf, &inList, &listType)
			hText := strings.TrimPrefix(line, "## ")
			hID := slugify(hText)
			toc = append(toc, TOCItem{ID: hID, Text: hText, Level: 2})
			bodyBuf.WriteString(fmt.Sprintf("<h2 id=\"%s\">%s</h2>\n", hID, parseInline(hText)))
			continue
		}

		if strings.HasPrefix(line, "### ") {
			closeList(&bodyBuf, &inList, &listType)
			hText := strings.TrimPrefix(line, "### ")
			hID := slugify(hText)
			toc = append(toc, TOCItem{ID: hID, Text: hText, Level: 3})
			bodyBuf.WriteString(fmt.Sprintf("<h3 id=\"%s\">%s</h3>\n", hID, parseInline(hText)))
			continue
		}

		// Lists
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			itemText := strings.TrimPrefix(strings.TrimPrefix(trimmed, "- "), "* ")
			if !inList || listType != "ul" {
				closeList(&bodyBuf, &inList, &listType)
				inList = true
				listType = "ul"
				bodyBuf.WriteString("<ul>\n")
			}
			bodyBuf.WriteString(fmt.Sprintf("  <li>%s</li>\n", parseInline(itemText)))
			continue
		}

		matched, _ := regexp.MatchString(`^\d+\.\s`, trimmed)
		if matched {
			re := regexp.MustCompile(`^\d+\.\s`)
			itemText := re.ReplaceAllString(trimmed, "")
			if !inList || listType != "ol" {
				closeList(&bodyBuf, &inList, &listType)
				inList = true
				listType = "ol"
				bodyBuf.WriteString("<ol>\n")
			}
			bodyBuf.WriteString(fmt.Sprintf("  <li>%s</li>\n", parseInline(itemText)))
			continue
		}

		// Horizontal rule
		if trimmed == "---" || trimmed == "***" {
			closeList(&bodyBuf, &inList, &listType)
			bodyBuf.WriteString("<hr>\n")
			continue
		}

		// Paragraph
		if trimmed != "" {
			closeList(&bodyBuf, &inList, &listType)
			bodyBuf.WriteString(fmt.Sprintf("<p>%s</p>\n", parseInline(line)))
		}
	}

	closeList(&bodyBuf, &inList, &listType)
	closeTable(&bodyBuf, &inTable, &tableLines)
	closeBlockquote(&bodyBuf, &inBlockquote)

	// Build TOC block if headings exist
	var tocBuf bytes.Buffer
	if len(toc) > 0 && !isIndex {
		tocBuf.WriteString("<div class=\"toc-box\">\n")
		tocBuf.WriteString("  <div class=\"toc-title\">Table of Contents</div>\n")
		tocBuf.WriteString("  <ul class=\"toc-list\">\n")
		for _, item := range toc {
			indent := ""
			if item.Level == 3 {
				indent = "&nbsp;&nbsp;↳ "
			}
			tocBuf.WriteString(fmt.Sprintf("    <li><a href=\"#%s\">%s%s</a></li>\n", item.ID, indent, html.EscapeString(item.Text)))
		}
		tocBuf.WriteString("  </ul>\n")
		tocBuf.WriteString("</div>\n")
	}

	// Back Link for non-index pages
	backLink := ""
	if !isIndex {
		backLink = "<a class=\"back-link\" href=\"/docs\">← Back to Documentation Home</a>\n"
	}

	// Active link helpers
	activeHome := ""
	if isIndex {
		activeHome = "class=\"active\""
	}

	// Full HTML page wrapper
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>%s - Go DDD Docs</title>
  <link rel="stylesheet" href="/docs/assets/docs.css">
</head>
<body>
  <header class="docs-header">
    <a href="/docs" class="docs-brand">
      <svg viewBox="0 0 24 24"><path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"/></svg>
      <span>go-ddd Docs</span>
    </a>
    <nav class="docs-nav">
      <a href="/docs" %s>Home</a>
      <a href="/docs?page=onboarding">Onboarding</a>
      <a href="/docs?page=architecture">Architecture</a>
      <a href="/docs?page=user-module">User Module</a>
      <a href="/swagger/" target="_blank">Swagger UI ↗</a>
    </nav>
  </header>

  <div class="docs-container">
    <aside class="docs-sidebar">
      <h4>Core Docs</h4>
      <ul>
        <li><a href="/docs?page=onboarding">Onboarding Guide</a></li>
        <li><a href="/docs?page=architecture">Architecture Reference</a></li>
      </ul>
      <h4>Guides</h4>
      <ul>
        <li><a href="/docs?page=external-client-integration">Client Integration</a></li>
        <li><a href="/docs?page=error-builder">Error Handling</a></li>
      </ul>
      <h4>Modules</h4>
      <ul>
        <li><a href="/docs?page=user-module">User Module</a></li>
      </ul>
      <h4>Templates</h4>
      <ul>
        <li><a href="/docs?page=feature-template">Feature Template</a></li>
        <li><a href="/docs?page=guide-template">Guide Template</a></li>
        <li><a href="/docs?page=swagger-description-template">Swagger Template</a></li>
      </ul>
    </aside>

    <main class="docs-main">
      %s
      %s
      %s
    </main>
  </div>

  <footer class="docs-footer">
    <p>&copy; 2026 go-ddd Team. Markdown-first Documentation Architecture.</p>
  </footer>
</body>
</html>
`, html.EscapeString(pageTitle), activeHome, backLink, tocBuf.String(), bodyBuf.String())
}

func closeList(buf *bytes.Buffer, inList *bool, listType *string) {
	if *inList {
		buf.WriteString(fmt.Sprintf("</%s>\n", *listType))
		*inList = false
		*listType = ""
	}
}

func closeBlockquote(buf *bytes.Buffer, inBlockquote *bool) {
	if *inBlockquote {
		buf.WriteString("</blockquote>\n")
		*inBlockquote = false
	}
}

func closeTable(buf *bytes.Buffer, inTable *bool, tableLines *[]string) {
	if !*inTable || len(*tableLines) == 0 {
		return
	}

	buf.WriteString("<table>\n")
	for i, line := range *tableLines {
		parts := strings.Split(line, "|")
		if len(parts) < 2 {
			continue
		}
		// Trim outer empty split parts
		var cells []string
		for idx, p := range parts {
			if idx == 0 && strings.TrimSpace(p) == "" {
				continue
			}
			if idx == len(parts)-1 && strings.TrimSpace(p) == "" {
				continue
			}
			cells = append(cells, strings.TrimSpace(p))
		}

		// Check divider row
		if i == 1 && isTableDivider(cells) {
			continue
		}

		tag := "td"
		if i == 0 {
			tag = "th"
			buf.WriteString("  <thead>\n  <tr>\n")
		} else if i == 1 || (i == 2 && isTableDivider(strings.Split((*tableLines)[1], "|"))) {
			if i == 1 || i == 2 {
				buf.WriteString("  <tbody>\n")
			}
			buf.WriteString("  <tr>\n")
		} else {
			buf.WriteString("  <tr>\n")
		}

		for _, cell := range cells {
			buf.WriteString(fmt.Sprintf("    <%s>%s</%s>\n", tag, parseInline(cell), tag))
		}
		buf.WriteString("  </tr>\n")

		if i == 0 {
			buf.WriteString("  </thead>\n")
		}
	}
	buf.WriteString("  </tbody>\n")
	buf.WriteString("</table>\n")

	*inTable = false
	*tableLines = nil
}

func isTableDivider(cells []string) bool {
	for _, c := range cells {
		trimmed := strings.ReplaceAll(strings.TrimSpace(c), "-", "")
		trimmed = strings.ReplaceAll(trimmed, ":", "")
		if trimmed != "" {
			return false
		}
	}
	return true
}

func parseInline(text string) string {
	// Escape HTML
	escaped := html.EscapeString(text)

	// Bold **text**
	reBold := regexp.MustCompile(`\*\*(.*?)\*\*`)
	escaped = reBold.ReplaceAllString(escaped, "<strong>$1</strong>")

	// Italic *text*
	reItalic := regexp.MustCompile(`\*(.*?)\*`)
	escaped = reItalic.ReplaceAllString(escaped, "<em>$1</em>")

	// Inline code `code`
	reCode := regexp.MustCompile("`([^`]+)`")
	escaped = reCode.ReplaceAllString(escaped, "<code>$1</code>")

	// Links [label](url)
	reLink := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	escaped = reLink.ReplaceAllString(escaped, `<a href="$2">$1</a>`)

	return escaped
}

func slugify(text string) string {
	slug := strings.ToLower(strings.TrimSpace(text))
	re := regexp.MustCompile(`[^a-z0-9\s-]`)
	slug = re.ReplaceAllString(slug, "")
	reSpaces := regexp.MustCompile(`[\s-]+`)
	slug = reSpaces.ReplaceAllString(slug, "-")
	return slug
}
