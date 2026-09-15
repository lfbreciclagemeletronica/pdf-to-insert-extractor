// Package cmd contém a definição dos comandos da CLI pdftoinsert.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/gstundner/pdftoinsert/internal/extractor"
	"github.com/gstundner/pdftoinsert/internal/parser"
	"github.com/gstundner/pdftoinsert/internal/sqlgen"
	"github.com/gstundner/pdftoinsert/internal/tui"
	"github.com/gstundner/pdftoinsert/internal/writer"
)

var (
	inputFile  string
	outputFile string
	viewFile   string
)

// rootCmd é o comando raiz da CLI pdftoinsert.
var rootCmd = &cobra.Command{
	Use:   "pdftoinsert",
	Short: "Extrai dados de PDFs do LFB e gera comandos SQL de INSERT",
	Long: "pdftoinsert lê um PDF de resultado de pesagem e triagem, extrai os\n" +
		"dados necessários (como o Fornecedor) e gera um arquivo com os\n" +
		"comandos INSERT correspondentes.\n\n" +
		"Use -v/--view para abrir uma interface interativa de terminal e\n" +
		"revisar os comandos INSERT de um arquivo SQL já gerado.",
	RunE: run,
}

// Execute inicia a execução da CLI.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().StringVarP(&inputFile, "file", "f", "", "Caminho do arquivo PDF de entrada")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Caminho do arquivo SQL de saída")
	rootCmd.Flags().StringVarP(&viewFile, "view", "v", "", "Abre uma TUI para visualizar um arquivo SQL já gerado")
}

func run(cmd *cobra.Command, args []string) error {
	if viewFile != "" {
		return tui.Run(viewFile)
	}

	if inputFile == "" || outputFile == "" {
		return fmt.Errorf("as flags -f/--file e -o/--output são obrigatórias (ou use -v/--view para visualizar um arquivo SQL existente)")
	}

	text, err := extractor.ExtractText(inputFile)
	if err != nil {
		return fmt.Errorf("falha ao extrair texto do PDF: %w", err)
	}

	header, err := parser.ParseHeader(text)
	if err != nil {
		return fmt.Errorf("falha ao localizar dados do cabeçalho: %w", err)
	}

	itens, err := parser.ParseItems(text)
	if err != nil {
		return fmt.Errorf("falha ao localizar itens do recibo: %w", err)
	}

	codigoRecibo := sqlgen.NewCodigoRecibo()

	inserts := []string{
		sqlgen.GenerateFornecedorInsert(header.Fornecedor),
		sqlgen.GenerateReciboCompraInsert(
			header.Fornecedor,
			header.DataRecibo,
			codigoRecibo,
			header.ValorTotal,
			header.PesoTotalKg,
		),
	}

	for _, item := range itens {
		inserts = append(inserts,
			sqlgen.GenerateItemInsert(item.Nome),
			sqlgen.GenerateReciboCompraItemInsert(codigoRecibo, item.Nome, item.PesoKg, item.PrecoPorKg, item.Total),
		)
	}

	if err := writer.WriteInserts(outputFile, inserts); err != nil {
		return fmt.Errorf("falha ao gravar arquivo de saída: %w", err)
	}

	fmt.Printf("Fornecedor extraído: %s\n", header.Fornecedor)
	fmt.Printf("Peso total: %.3f kg\n", header.PesoTotalKg)
	fmt.Printf("Valor total: R$ %.2f\n", header.ValorTotal)
	fmt.Printf("Data do recibo: %s\n", header.DataRecibo)
	fmt.Printf("Código do recibo: %s\n", codigoRecibo)
	fmt.Printf("Itens extraídos: %d\n", len(itens))
	fmt.Printf("INSERT(s) gravado(s) em: %s\n", outputFile)

	return nil
}
