// Package extractor lida com a extração de texto plano de arquivos PDF.
package extractor

import (
	"fmt"
	"strings"

	"github.com/ledongthuc/pdf"
)

// ExtractText lê o arquivo PDF em pdfPath e retorna todo o texto plano
// concatenado de todas as páginas.
func ExtractText(pdfPath string) (string, error) {
	f, r, err := pdf.Open(pdfPath)
	if err != nil {
		return "", fmt.Errorf("abrir PDF %q: %w", pdfPath, err)
	}
	defer f.Close()

	var sb strings.Builder
	totalPages := r.NumPage()
	for pageIndex := 1; pageIndex <= totalPages; pageIndex++ {
		page := r.Page(pageIndex)
		if page.V.IsNull() {
			continue
		}

		text, err := page.GetPlainText(nil)
		if err != nil {
			return "", fmt.Errorf("extrair texto da página %d: %w", pageIndex, err)
		}
		sb.WriteString(text)
		sb.WriteString("\n")
	}

	return sb.String(), nil
}
