// Package parser extrai campos estruturados do texto plano de um PDF.
package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Header representa os dados do cabeçalho do relatório "RESULTADO DA
// PESAGEM E TRIAGEM LFB", extraídos do texto do PDF.
//
// O layout do PDF possui duas variações conhecidas:
//   - modelo antigo: rótulos "PESO TOTAL" e "VALOR TOTAL";
//   - modelo novo: rótulos "PESO" e "VALOR".
//
// Ambas são suportadas pelo parser.
type Header struct {
	// Fornecedor é o nome do fornecedor (também usado como Nome do Usuário).
	Fornecedor string
	// PesoTotalKg é o peso total em quilogramas.
	PesoTotalKg float64
	// ValorTotal é o valor total em reais.
	ValorTotal float64
	// DataRecibo é a data do recibo, no formato dd/mm/aaaa exatamente como
	// aparece no PDF.
	DataRecibo string
}

// whitespaceRegex identifica sequências de espaços em branco (incluindo
// quebras de linha), usado para normalizar textos capturados que podem
// estar quebrados em múltiplas linhas no PDF (ex.: nomes longos de
// fornecedor que quebram a linha na tabela).
var whitespaceRegex = regexp.MustCompile(`\s+`)

// fornecedorRegex captura o nome do fornecedor entre o marcador "FORNECEDOR"
// e o próximo marcador conhecido de peso ("PESO" ou "PESO TOTAL"). A flag
// (?s) permite que "." cruze quebras de linha, necessário pois o nome do
// fornecedor às vezes é quebrado em múltiplas linhas pela tabela do PDF.
var fornecedorRegex = regexp.MustCompile(`(?is)FORNECEDOR\s+(.+?)\s+PESO(?:\s+TOTAL)?\b`)

// pesoRegex captura o valor numérico do peso total (em kg), aceitando os
// rótulos "PESO" ou "PESO TOTAL".
var pesoRegex = regexp.MustCompile(`(?is)PESO(?:\s+TOTAL)?\s+([\d.,]+)\s*kg`)

// valorRegex captura o valor numérico do valor total (em R$), aceitando os
// rótulos "VALOR" ou "VALOR TOTAL".
var valorRegex = regexp.MustCompile(`(?is)VALOR(?:\s+TOTAL)?\s+R\$\s*([\d.,]+)`)

// dataRegex captura a data do recibo no formato dd/mm/aaaa.
var dataRegex = regexp.MustCompile(`(?is)DATA\s+(\d{2}/\d{2}/\d{4})`)

// ParseFornecedor localiza e retorna o nome do fornecedor a partir do texto
// extraído de um PDF. Retorna erro caso o marcador "FORNECEDOR" não seja
// encontrado.
func ParseFornecedor(text string) (string, error) {
	matches := fornecedorRegex.FindStringSubmatch(text)
	if len(matches) < 2 {
		return "", fmt.Errorf("campo FORNECEDOR não encontrado no texto do PDF")
	}

	fornecedor := normalizeWhitespace(matches[1])
	if fornecedor == "" {
		return "", fmt.Errorf("campo FORNECEDOR encontrado, mas vazio")
	}

	return fornecedor, nil
}

// ParseHeader localiza e retorna todos os dados do cabeçalho do relatório
// (Fornecedor, PesoTotalKg, ValorTotal e DataRecibo) a partir do texto
// extraído de um PDF.
func ParseHeader(text string) (Header, error) {
	fornecedor, err := ParseFornecedor(text)
	if err != nil {
		return Header{}, err
	}

	pesoMatch := pesoRegex.FindStringSubmatch(text)
	if len(pesoMatch) < 2 {
		return Header{}, fmt.Errorf("campo PESO não encontrado no texto do PDF")
	}
	pesoTotalKg, err := parseBRFloat(pesoMatch[1])
	if err != nil {
		return Header{}, fmt.Errorf("converter PESO %q: %w", pesoMatch[1], err)
	}

	valorMatch := valorRegex.FindStringSubmatch(text)
	if len(valorMatch) < 2 {
		return Header{}, fmt.Errorf("campo VALOR não encontrado no texto do PDF")
	}
	valorTotal, err := parseBRFloat(valorMatch[1])
	if err != nil {
		return Header{}, fmt.Errorf("converter VALOR %q: %w", valorMatch[1], err)
	}

	dataMatch := dataRegex.FindStringSubmatch(text)
	if len(dataMatch) < 2 {
		return Header{}, fmt.Errorf("campo DATA não encontrado no texto do PDF")
	}

	return Header{
		Fornecedor:  fornecedor,
		PesoTotalKg: pesoTotalKg,
		ValorTotal:  valorTotal,
		DataRecibo:  dataMatch[1],
	}, nil
}

// normalizeWhitespace colapsa qualquer sequência de espaços em branco
// (incluindo quebras de linha) em um único espaço, e remove espaços nas
// extremidades.
func normalizeWhitespace(s string) string {
	return strings.TrimSpace(whitespaceRegex.ReplaceAllString(s, " "))
}

// parseBRFloat converte um número no formato brasileiro (ponto como
// separador de milhar, vírgula como separador decimal) para float64.
func parseBRFloat(s string) (float64, error) {
	normalized := strings.ReplaceAll(s, ".", "")
	normalized = strings.ReplaceAll(normalized, ",", ".")
	return strconv.ParseFloat(normalized, 64)
}
