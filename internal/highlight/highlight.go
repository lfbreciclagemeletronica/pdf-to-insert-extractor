// Package highlight aplica coloração de sintaxe SQL usando chroma, para uso
// em interfaces de terminal (TUI).
package highlight

import (
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// SQL recebe o conteúdo completo de um arquivo SQL e retorna suas linhas já
// coloridas com sequências ANSI, uma entrada por linha do arquivo original.
func SQL(code string) ([]string, error) {
	lexer := lexers.Get("sql")
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}

	formatter := formatters.Get("terminal16m")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return nil, err
	}

	var sb strings.Builder
	if err := formatter.Format(&sb, style, iterator); err != nil {
		return nil, err
	}

	highlighted := sb.String()
	// chroma sempre finaliza com uma quebra de linha; removemos para que a
	// divisão abaixo produza o mesmo número de linhas do arquivo original.
	highlighted = strings.TrimSuffix(highlighted, "\n")

	return strings.Split(highlighted, "\n"), nil
}
