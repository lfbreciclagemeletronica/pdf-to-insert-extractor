package sqlgen

import (
	"regexp"
	"strings"
	"testing"
)

func TestGenerateFornecedorInsert(t *testing.T) {
	got := GenerateFornecedorInsert("Maurício Amaral da Costa")

	wantContains := []string{
		"INSERT INTO Usuarios (Nome, Cpf, Cnpj, ChavePix)",
		"VALUES ('Maurício Amaral da Costa', NULL, NULL, NULL)",
		"ON CONFLICT(Nome) DO NOTHING;",
	}

	for _, want := range wantContains {
		if !strings.Contains(got, want) {
			t.Fatalf("esperava que o SQL contivesse %q, obteve:\n%s", want, got)
		}
	}
}

func TestGenerateFornecedorInsert_EscapeAspasSimples(t *testing.T) {
	got := GenerateFornecedorInsert("O'Brien")

	if !strings.Contains(got, "O''Brien") {
		t.Fatalf("esperava aspas simples escapadas, obteve:\n%s", got)
	}
}

// codigoReciboRegex valida que um valor gerado por NewCodigoRecibo é uma
// string contendo exatamente 5 dígitos (ex.: "12345").
var codigoReciboRegex = regexp.MustCompile(`^\d{5}$`)

func TestNewCodigoRecibo(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 20; i++ {
		codigo := NewCodigoRecibo()
		if !codigoReciboRegex.MatchString(codigo) {
			t.Fatalf("esperava um código de 5 dígitos, obteve: %q", codigo)
		}
		seen[codigo] = true
	}

	// Com 20 gerações aleatórias entre 10000-99999, é extremamente
	// improvável (mas não impossível) obter o mesmo valor sempre; exigimos
	// ao menos alguma variação para garantir que não é um valor fixo.
	if len(seen) <= 1 {
		t.Fatalf("esperava CodigoRecibo variando entre execuções, obteve sempre o mesmo valor: %v", seen)
	}
}

func TestGenerateReciboCompraInsert(t *testing.T) {
	got := GenerateReciboCompraInsert("Gabriel Fanto", "26/04/1996", "12345", 123.123, 12.00)

	wantContains := []string{
		"INSERT INTO RecibosCompra (UsuarioId, DataRecibo, CodigoRecibo, ValorTotal, PesoTotalKg)",
		"VALUES ((SELECT Id FROM Usuarios WHERE Nome = 'Gabriel Fanto'), '26/04/1996', '12345', 123.123, 12);",
	}

	for _, want := range wantContains {
		if !strings.Contains(got, want) {
			t.Fatalf("esperava que o SQL contivesse %q, obteve:\n%s", want, got)
		}
	}
}

func TestGenerateReciboCompraInsert_EscapeAspasSimples(t *testing.T) {
	got := GenerateReciboCompraInsert("O'Brien", "01/01/2026", "12345", 10, 1)

	if !strings.Contains(got, "O''Brien") {
		t.Fatalf("esperava aspas simples escapadas, obteve:\n%s", got)
	}
}

func TestGenerateItemInsert(t *testing.T) {
	got := GenerateItemInsert("Placa Drive")

	want := "INSERT INTO Itens (Nome)\n" +
		"VALUES ('Placa Drive')\n" +
		"ON CONFLICT(Nome) DO NOTHING;"

	if got != want {
		t.Fatalf("GenerateItemInsert() =\n%s\nwant:\n%s", got, want)
	}
}

func TestGenerateItemInsert_EscapeAspasSimples(t *testing.T) {
	got := GenerateItemInsert("O'Brien")

	if !strings.Contains(got, "O''Brien") {
		t.Fatalf("esperava aspas simples escapadas, obteve:\n%s", got)
	}
}

func TestGenerateReciboCompraItemInsert(t *testing.T) {
	got := GenerateReciboCompraItemInsert("12345", "Placa Drive", 1.412, 42.00, 59.30)

	want := "INSERT INTO ReciboCompraItens (ReciboCodigo, ItemId, PesoKg, PrecoPorKg, Total)\n" +
		"VALUES (12345, (SELECT Id FROM Itens WHERE Nome = 'Placa Drive'), 1.412, 42, 59.3);"

	if got != want {
		t.Fatalf("GenerateReciboCompraItemInsert() =\n%s\nwant:\n%s", got, want)
	}
}

func TestGenerateReciboCompraItemInsert_EscapeAspasSimples(t *testing.T) {
	got := GenerateReciboCompraItemInsert("12345", "O'Brien", 1, 1, 1)

	if !strings.Contains(got, "O''Brien") {
		t.Fatalf("esperava aspas simples escapadas, obteve:\n%s", got)
	}
}
