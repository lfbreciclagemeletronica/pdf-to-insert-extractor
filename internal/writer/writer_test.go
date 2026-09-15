package writer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteInserts_CriaArquivoNovo(t *testing.T) {
	path := filepath.Join(t.TempDir(), "output.sql")

	if err := WriteInserts(path, []string{"INSERT INTO A VALUES (1);", "INSERT INTO B VALUES (2);"}); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("erro ao ler arquivo: %v", err)
	}

	want := "INSERT INTO A VALUES (1);\n\nINSERT INTO B VALUES (2);\n"
	if string(got) != want {
		t.Fatalf("conteúdo = %q, want %q", string(got), want)
	}
}

func TestWriteInserts_AcrescentaAoArquivoExistente(t *testing.T) {
	path := filepath.Join(t.TempDir(), "output.sql")

	if err := WriteInserts(path, []string{"INSERT INTO A VALUES (1);"}); err != nil {
		t.Fatalf("erro na primeira escrita: %v", err)
	}

	if err := WriteInserts(path, []string{"INSERT INTO B VALUES (2);"}); err != nil {
		t.Fatalf("erro na segunda escrita: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("erro ao ler arquivo: %v", err)
	}

	want := "INSERT INTO A VALUES (1);\n\nINSERT INTO B VALUES (2);\n"
	if string(got) != want {
		t.Fatalf("conteúdo = %q, want %q", string(got), want)
	}
}

func TestWriteInserts_NaoRemoveConteudoManualExistente(t *testing.T) {
	path := filepath.Join(t.TempDir(), "output.sql")

	if err := os.WriteFile(path, []byte("-- comentário existente\n"), 0o644); err != nil {
		t.Fatalf("erro ao preparar arquivo: %v", err)
	}

	if err := WriteInserts(path, []string{"INSERT INTO A VALUES (1);"}); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("erro ao ler arquivo: %v", err)
	}

	want := "-- comentário existente\n\nINSERT INTO A VALUES (1);\n"
	if string(got) != want {
		t.Fatalf("conteúdo = %q, want %q", string(got), want)
	}
}
