// Package writer grava os comandos SQL gerados em um arquivo de saída.
package writer

import (
	"fmt"
	"os"
	"strings"
)

// WriteInserts grava a lista de comandos INSERT no arquivo outputPath,
// separados por uma linha em branco. O conteúdo é acrescentado ao final do
// arquivo (append) caso ele já exista, permitindo extrair múltiplos PDFs
// para o mesmo arquivo SQL sem perder o que já havia sido gravado. Caso o
// arquivo não exista, ele é criado.
func WriteInserts(outputPath string, inserts []string) error {
	content := strings.Join(inserts, "\n\n") + "\n"

	f, err := os.OpenFile(outputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("abrir arquivo de saída %q: %w", outputPath, err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return fmt.Errorf("verificar arquivo de saída %q: %w", outputPath, err)
	}

	// Se o arquivo já possui conteúdo, garante uma linha em branco de
	// separação antes de acrescentar os novos INSERTs.
	if info.Size() > 0 {
		content = "\n" + content
	}

	if _, err := f.WriteString(content); err != nil {
		return fmt.Errorf("escrever arquivo de saída %q: %w", outputPath, err)
	}

	return nil
}
