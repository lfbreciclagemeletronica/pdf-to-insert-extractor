// Command pdftoinsert extrai dados de PDFs e gera comandos SQL de INSERT.
package main

import (
	"fmt"
	"os"

	"github.com/gstundner/pdftoinsert/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(1)
	}
}
