package parser

import (
	"testing"

	"github.com/gstundner/pdftoinsert/internal/extractor"
)

const textModeloNovoInline = "RESULTADO DA PESAGEM E TRIAGEM LFB\n" +
	"FORNECEDOR  Mauricio Amaral da Costa PESO 13,134 kg VALOR R$ 394,53 DATA 15/05/2026\n"

// textModeloAntigo reproduz o texto extraído de um PDF no modelo antigo
// (rótulos "PESO TOTAL"/"VALOR TOTAL", cada campo em sua própria linha).
const textModeloAntigo = "LFB RECICLAGEM ELETRONICA\n" +
	"RESULTADO DA PESAGEM E TRIAGEM LFB\n" +
	"FORNECEDOR\n" +
	"Rafael Henrique koelling Avila\n" +
	"PESO TOTAL\n" +
	"5,098 kg\n" +
	"VALOR TOTAL\n" +
	"R$ 379,99\n" +
	"DATA\n" +
	"09/04/2026\n" +
	"KG\nVALOR\nTOTAL\n"

// textModeloNovoMultilinha reproduz o texto extraído de um PDF no modelo
// novo (rótulos "PESO"/"VALOR") em que o nome do fornecedor quebra em duas
// linhas por ser longo demais para a célula da tabela.
const textModeloNovoMultilinha = "LFB RECICLAGEM ELETRONICA\n" +
	"RESULTADO DA PESAGEM E TRIAGEM LFB\n" +
	"FORNECEDOR\n" +
	"ALESSANDRO CARDOSO\n" +
	"BOEIRA\n" +
	"PESO\n" +
	"3,856 kg\n" +
	"VALOR\n" +
	"R$ 24,78\n" +
	"DATA\n" +
	"08/09/2026\n" +
	"MATERIAL\nKG\nVALOR/KG\nTOTAL\n"

func TestParseFornecedor(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		want    string
		wantErr bool
	}{
		{
			name: "modelo novo inline",
			text: textModeloNovoInline,
			want: "Mauricio Amaral da Costa",
		},
		{
			name: "modelo antigo",
			text: textModeloAntigo,
			want: "Rafael Henrique koelling Avila",
		},
		{
			name: "modelo novo, nome em duas linhas",
			text: textModeloNovoMultilinha,
			want: "ALESSANDRO CARDOSO BOEIRA",
		},
		{
			name:    "marcador ausente",
			text:    "algum texto sem o campo esperado",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFornecedor(tt.text)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("esperava erro, obteve nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("erro inesperado: %v", err)
			}
			if got != tt.want {
				t.Fatalf("fornecedor = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseHeader(t *testing.T) {
	tests := []struct {
		name string
		text string
		want Header
	}{
		{
			name: "modelo novo inline",
			text: textModeloNovoInline,
			want: Header{
				Fornecedor:  "Mauricio Amaral da Costa",
				PesoTotalKg: 13.134,
				ValorTotal:  394.53,
				DataRecibo:  "15/05/2026",
			},
		},
		{
			name: "modelo antigo",
			text: textModeloAntigo,
			want: Header{
				Fornecedor:  "Rafael Henrique koelling Avila",
				PesoTotalKg: 5.098,
				ValorTotal:  379.99,
				DataRecibo:  "09/04/2026",
			},
		},
		{
			name: "modelo novo, nome em duas linhas",
			text: textModeloNovoMultilinha,
			want: Header{
				Fornecedor:  "ALESSANDRO CARDOSO BOEIRA",
				PesoTotalKg: 3.856,
				ValorTotal:  24.78,
				DataRecibo:  "08/09/2026",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseHeader(tt.text)
			if err != nil {
				t.Fatalf("erro inesperado: %v", err)
			}
			if got != tt.want {
				t.Fatalf("ParseHeader() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// TestParseHeader_ArquivosReais garante que o parser funciona com os PDFs de
// teste reais do projeto, extraindo o texto via internal/extractor.
func TestParseHeader_ArquivosReais(t *testing.T) {
	files := []string{
		"../../testdata/exemplo.pdf",
		"../../testdata/Rafael_Henrique_koelling_Avila_09-04-2026.pdf",
		"../../testdata/ALESSANDRO_CARDOSO_BOEIRA_08092026143807.pdf",
	}

	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			text, err := extractor.ExtractText(file)
			if err != nil {
				t.Fatalf("erro ao extrair texto do PDF de teste: %v", err)
			}

			header, err := ParseHeader(text)
			if err != nil {
				t.Fatalf("erro ao parsear cabeçalho: %v", err)
			}

			if header.Fornecedor == "" {
				t.Errorf("fornecedor está vazio")
			}
			if header.PesoTotalKg <= 0 {
				t.Errorf("PesoTotalKg = %v, esperava > 0", header.PesoTotalKg)
			}
			if header.ValorTotal <= 0 {
				t.Errorf("ValorTotal = %v, esperava > 0", header.ValorTotal)
			}
			if header.DataRecibo == "" {
				t.Errorf("DataRecibo está vazio")
			}

			t.Logf("header extraído: %+v", header)
		})
	}
}
