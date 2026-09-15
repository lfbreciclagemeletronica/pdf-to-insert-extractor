package sqlparser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	content := `-- comentário
INSERT INTO Usuarios (Nome, Cpf, Cnpj, ChavePix)
VALUES ('Maurício Amaral da Costa', NULL, NULL, NULL)
ON CONFLICT(Nome) DO NOTHING;

INSERT INTO Materiais (Nome, ValorKg)
VALUES ('Placa Notebook C', 70.00);
`
	dir := t.TempDir()
	path := filepath.Join(dir, "output.sql")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("erro ao criar arquivo de teste: %v", err)
	}

	file, err := Load(path)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if len(file.Inserts) != 2 {
		t.Fatalf("esperava 2 inserts, obteve %d", len(file.Inserts))
	}

	if file.Inserts[0].Table != "Usuarios" {
		t.Errorf("insert[0].Table = %q, want %q", file.Inserts[0].Table, "Usuarios")
	}
	if file.Inserts[0].Line != 2 {
		t.Errorf("insert[0].Line = %d, want %d", file.Inserts[0].Line, 2)
	}
	if file.Inserts[1].Table != "Materiais" {
		t.Errorf("insert[1].Table = %q, want %q", file.Inserts[1].Table, "Materiais")
	}
	if file.Inserts[1].Line != 6 {
		t.Errorf("insert[1].Line = %d, want %d", file.Inserts[1].Line, 6)
	}

	wantTables := []string{"Usuarios", "Materiais"}
	if len(file.Tables) != len(wantTables) {
		t.Fatalf("Tables = %v, want %v", file.Tables, wantTables)
	}
	for i, want := range wantTables {
		if file.Tables[i] != want {
			t.Errorf("Tables[%d] = %q, want %q", i, file.Tables[i], want)
		}
	}
}

func TestLoad_ArquivoInexistente(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "nao-existe.sql"))
	if err == nil {
		t.Fatal("esperava erro para arquivo inexistente")
	}
}
