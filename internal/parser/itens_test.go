package parser

import (
	"testing"

	"github.com/gstundner/pdftoinsert/internal/extractor"
)

func TestParseItems(t *testing.T) {
	text := textModeloAntigo + "\n" +
		"Placa Mãe C\n1,412\nR$ 42,00\nR$ 59,30\n" +
		"Placa Mãe D\n0,458\nR$ 20,00\nR$ 9,16\n"

	items, err := ParseItems(text)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	want := []Item{
		{Nome: "Placa Mãe C", PesoKg: 1.412, PrecoPorKg: 42.00, Total: 59.30},
		{Nome: "Placa Mãe D", PesoKg: 0.458, PrecoPorKg: 20.00, Total: 9.16},
	}

	if len(items) != len(want) {
		t.Fatalf("len(items) = %d, want %d (items: %+v)", len(items), len(want), items)
	}
	for i, w := range want {
		if items[i] != w {
			t.Errorf("items[%d] = %+v, want %+v", i, items[i], w)
		}
	}
}

func TestParseItems_NomeComPontuacao(t *testing.T) {
	text := textModeloNovoMultilinha + "\n" +
		"Desmanche Eletrônicos Consultar\n1,984\nR$ 1,50\nR$ 2,98\n" +
		"OUTROS\n1,026\nR$ 4,00\nR$ 4,10\n"

	items, err := ParseItems(text)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	want := []Item{
		{Nome: "Desmanche Eletrônicos Consultar", PesoKg: 1.984, PrecoPorKg: 1.50, Total: 2.98},
		{Nome: "OUTROS", PesoKg: 1.026, PrecoPorKg: 4.00, Total: 4.10},
	}

	if len(items) != len(want) {
		t.Fatalf("len(items) = %d, want %d (items: %+v)", len(items), len(want), items)
	}
	for i, w := range want {
		if items[i] != w {
			t.Errorf("items[%d] = %+v, want %+v", i, items[i], w)
		}
	}
}

func TestParseItems_CabecalhoAusente(t *testing.T) {
	_, err := ParseItems("texto sem tabela de materiais")
	if err == nil {
		t.Fatal("esperava erro para texto sem cabeçalho de materiais")
	}
}

// TestParseItems_ArquivosReais garante que o parser funciona com os PDFs de
// teste reais do projeto.
func TestParseItems_ArquivosReais(t *testing.T) {
	files := map[string]int{
		"../../testdata/exemplo.pdf":                                  13,
		"../../testdata/Rafael_Henrique_koelling_Avila_09-04-2026.pdf": 10,
		"../../testdata/ALESSANDRO_CARDOSO_BOEIRA_08092026143807.pdf":  4,
	}

	for file, wantCount := range files {
		t.Run(file, func(t *testing.T) {
			text, err := extractor.ExtractText(file)
			if err != nil {
				t.Fatalf("erro ao extrair texto do PDF de teste: %v", err)
			}

			items, err := ParseItems(text)
			if err != nil {
				t.Fatalf("erro ao parsear itens: %v", err)
			}

			if len(items) != wantCount {
				t.Errorf("len(items) = %d, want %d", len(items), wantCount)
			}

			for _, item := range items {
				if item.Nome == "" {
					t.Errorf("item com nome vazio: %+v", item)
				}
				if item.PesoKg <= 0 {
					t.Errorf("item %q com PesoKg <= 0: %+v", item.Nome, item)
				}
			}

			t.Logf("itens extraídos: %+v", items)
		})
	}
}
