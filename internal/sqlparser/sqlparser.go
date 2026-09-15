// Package sqlparser localiza comandos INSERT dentro de um arquivo SQL para
// fins de navegação em uma interface TUI.
package sqlparser

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Insert representa um comando INSERT localizado no arquivo SQL.
type Insert struct {
	// Table é o nome da tabela alvo do INSERT (ex.: "Usuarios").
	Table string
	// Line é o número da linha (1-based) onde o INSERT começa no arquivo.
	Line int
	// Summary é uma descrição curta usada para exibição na lista (ex.:
	// primeiro valor literal encontrado no comando).
	Summary string
}

// insertRegex captura o nome da tabela em uma linha "INSERT INTO <tabela>".
var insertRegex = regexp.MustCompile(`(?i)^\s*INSERT\s+INTO\s+([A-Za-z0-9_"` + "`" + `\[\]\.]+)`)

// valueRegex captura o primeiro literal de string em uma cláusula VALUES,
// usado apenas para compor um resumo legível na lista de INSERTs.
var valueRegex = regexp.MustCompile(`'([^']*)'`)

// File representa o conteúdo de um arquivo SQL já carregado, junto dos
// INSERTs localizados e das tabelas distintas envolvidas.
type File struct {
	// Lines contém todas as linhas do arquivo, na ordem original.
	Lines []string
	// Inserts contém todos os comandos INSERT encontrados, na ordem em que aparecem.
	Inserts []Insert
	// Tables contém os nomes distintos de tabelas afetadas, na ordem de primeira aparição.
	Tables []string
}

// Load lê o arquivo em path e localiza todos os comandos INSERT presentes.
func Load(path string) (*File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("abrir arquivo SQL %q: %w", path, err)
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("ler arquivo SQL %q: %w", path, err)
	}

	result := &File{Lines: lines}
	seenTables := map[string]bool{}

	for i, line := range lines {
		matches := insertRegex.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		table := strings.Trim(matches[1], `"`+"`"+`[]`)

		summary := table
		// A cláusula VALUES pode estar na própria linha do INSERT INTO ou em
		// uma das linhas seguintes (o gerador de SQL do projeto quebra em
		// múltiplas linhas), então procuramos nas próximas linhas também.
		for lookahead := i; lookahead < len(lines) && lookahead < i+4; lookahead++ {
			if vm := valueRegex.FindStringSubmatch(lines[lookahead]); vm != nil {
				summary = fmt.Sprintf("%s: %s", table, vm[1])
				break
			}
		}

		result.Inserts = append(result.Inserts, Insert{
			Table:   table,
			Line:    i + 1,
			Summary: summary,
		})

		if !seenTables[table] {
			seenTables[table] = true
			result.Tables = append(result.Tables, table)
		}
	}

	return result, nil
}
