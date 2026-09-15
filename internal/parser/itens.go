package parser

import (
	"fmt"
	"regexp"
	"strings"
)

// Item representa uma linha da tabela de materiais do relatório
// "RESULTADO DA PESAGEM E TRIAGEM LFB".
type Item struct {
	// Nome é o nome do material/item.
	Nome string
	// PesoKg é o peso do item em quilogramas.
	PesoKg float64
	// PrecoPorKg é o preço pago por quilograma (em reais).
	PrecoPorKg float64
	// Total é o valor total pago pelo item (em reais).
	Total float64
}

// itemsHeaderRegex localiza a linha "TOTAL" que encerra o cabeçalho da
// tabela de materiais (ex.: "MATERIAL\nKG\nVALOR/KG\nTOTAL" no modelo novo,
// ou "KG\nVALOR\nTOTAL" no modelo antigo). Os itens começam logo após essa
// linha.
var itemsHeaderRegex = regexp.MustCompile(`(?im)^\s*TOTAL\s*$`)

// kgLineRegex identifica uma linha contendo apenas um número (o peso de um
// item, ex.: "0,262"), sem prefixo de moeda.
var kgLineRegex = regexp.MustCompile(`^\d+(?:[.,]\d+)*$`)

// moneyLineRegex identifica uma linha de valor monetário (ex.: "R$ 70,00").
var moneyLineRegex = regexp.MustCompile(`(?i)^R\$\s*[\d.,]+$`)

// ParseItems localiza e retorna todos os itens (materiais) listados no
// relatório, a partir do texto extraído do PDF.
//
// O parser percorre as linhas após o cabeçalho da tabela como uma máquina
// de estados: acumula linhas de texto como nome do item até encontrar uma
// linha puramente numérica (peso em kg) seguida de duas linhas de valores
// monetários (preço por kg e total), que fecham o registro do item.
func ParseItems(text string) ([]Item, error) {
	loc := itemsHeaderRegex.FindStringIndex(text)
	if loc == nil {
		return nil, fmt.Errorf("cabeçalho da tabela de materiais não encontrado no texto do PDF")
	}

	var lines []string
	for _, rawLine := range strings.Split(text[loc[1]:], "\n") {
		line := strings.TrimSpace(rawLine)
		if line != "" {
			lines = append(lines, line)
		}
	}

	var items []Item
	var nameParts []string

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Uma linha de peso só marca o fim do nome do item se for seguida
		// por duas linhas de valores monetários (preço/kg e total).
		if kgLineRegex.MatchString(line) &&
			i+2 < len(lines) &&
			moneyLineRegex.MatchString(lines[i+1]) &&
			moneyLineRegex.MatchString(lines[i+2]) {

			nome := normalizeWhitespace(strings.Join(nameParts, " "))
			nameParts = nil

			if nome == "" {
				// Peso sem nome associado; ignora linha inesperada e
				// consome os valores monetários mesmo assim para não
				// desalinhar o parser.
				i += 2
				continue
			}

			pesoKg, err := parseBRFloat(line)
			if err != nil {
				return nil, fmt.Errorf("converter peso do item %q: %w", nome, err)
			}

			precoPorKg, err := parseBRFloat(extractMoneyAmount(lines[i+1]))
			if err != nil {
				return nil, fmt.Errorf("converter preço/kg do item %q: %w", nome, err)
			}

			total, err := parseBRFloat(extractMoneyAmount(lines[i+2]))
			if err != nil {
				return nil, fmt.Errorf("converter total do item %q: %w", nome, err)
			}

			items = append(items, Item{
				Nome:       nome,
				PesoKg:     pesoKg,
				PrecoPorKg: precoPorKg,
				Total:      total,
			})

			i += 2
			continue
		}

		nameParts = append(nameParts, line)
	}

	return items, nil
}

// extractMoneyAmount remove o prefixo "R$" e espaços de uma linha de valor
// monetário (ex.: "R$ 70,00" -> "70,00"), deixando o número pronto para
// parseBRFloat.
func extractMoneyAmount(line string) string {
	amount := strings.TrimSpace(line)
	amount = strings.TrimPrefix(amount, "R$")
	amount = strings.TrimPrefix(amount, "r$")
	return strings.TrimSpace(amount)
}
