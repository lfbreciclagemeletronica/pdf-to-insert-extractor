// Package sqlgen gera comandos SQL de INSERT a partir dos dados extraídos.
package sqlgen

import (
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"
)

const (
	// usuariosTable é o nome da tabela de destino para os dados de fornecedor.
	usuariosTable = "Usuarios"
	// recibosCompraTable é o nome da tabela de destino para os recibos de compra.
	recibosCompraTable = "RecibosCompra"
	// itensTable é o nome da tabela de destino para os itens/materiais.
	itensTable = "Itens"
	// reciboCompraItensTable é o nome da tabela de destino para os itens de
	// cada recibo de compra.
	reciboCompraItensTable = "ReciboCompraItens"

	// codigoReciboMin e codigoReciboMax delimitam a faixa de números de 5
	// dígitos (10000-99999) usada para gerar o CodigoRecibo.
	codigoReciboMin = 10000
	codigoReciboMax = 99999
)

// escapeSQLString escapa aspas simples para evitar quebra de sintaxe SQL.
func escapeSQLString(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

// GenerateFornecedorInsert gera o comando INSERT INTO para a tabela Usuarios
// a partir do nome do fornecedor extraído do PDF.
//
// Cpf e Cnpj não estão disponíveis no PDF e são gravados como NULL.
// ChavePix também é gravado como NULL, devendo ser preenchido manualmente depois.
//
// O INSERT usa ON CONFLICT(Nome) DO NOTHING para evitar duplicidade de forma
// nativa do SQLite. A estrutura pode ser alterada para DO UPDATE no futuro
// se necessário para atualizar registros existentes.
func GenerateFornecedorInsert(nomeFornecedor string) string {
	nome := escapeSQLString(nomeFornecedor)
	return fmt.Sprintf(
		"INSERT INTO %s (Nome, Cpf, Cnpj, ChavePix)\n"+
			"VALUES ('%s', NULL, NULL, NULL)\n"+
			"ON CONFLICT(Nome) DO NOTHING;",
		usuariosTable,
		nome,
	)
}

// NewCodigoRecibo gera um código aleatório de 5 dígitos (10000-99999) para
// ser usado como identificador de um recibo (CodigoRecibo em RecibosCompra
// e ReciboCodigo em ReciboCompraItens). Deve ser gerado uma única vez por
// recibo e reutilizado em todos os INSERTs relacionados a ele.
func NewCodigoRecibo() string {
	return strconv.Itoa(rand.IntN(codigoReciboMax-codigoReciboMin+1) + codigoReciboMin)
}

// GenerateReciboCompraInsert gera o comando INSERT INTO para a tabela
// RecibosCompra a partir dos dados de cabeçalho extraídos do PDF.
//
// O UsuarioId é resolvido em tempo de execução do SQL via subquery que busca
// o Id na tabela Usuarios pelo Nome do fornecedor, então o INSERT do
// fornecedor (GenerateFornecedorInsert) deve ser executado antes deste.
//
// codigoRecibo deve ser gerado previamente com NewCodigoRecibo e reutilizado
// nos INSERTs de ReciboCompraItens associados a este recibo.
func GenerateReciboCompraInsert(nomeFornecedor, dataRecibo, codigoRecibo string, valorTotal, pesoTotalKg float64) string {
	nome := escapeSQLString(nomeFornecedor)
	data := escapeSQLString(dataRecibo)
	return fmt.Sprintf(
		"INSERT INTO %s (UsuarioId, DataRecibo, CodigoRecibo, ValorTotal, PesoTotalKg)\n"+
			"VALUES ((SELECT Id FROM %s WHERE Nome = '%s'), '%s', '%s', %s, %s);",
		recibosCompraTable,
		usuariosTable,
		nome,
		data,
		escapeSQLString(codigoRecibo),
		formatFloat(valorTotal),
		formatFloat(pesoTotalKg),
	)
}

// GenerateItemInsert gera o comando INSERT INTO para a tabela Itens a
// partir do nome de um item/material extraído do PDF.
//
// O INSERT usa ON CONFLICT(Nome) DO NOTHING para evitar duplicidade,
// assumindo que a coluna Nome possui uma constraint UNIQUE (necessária para
// que o Id do item possa ser resolvido de forma determinística pelo nome).
func GenerateItemInsert(nomeItem string) string {
	nome := escapeSQLString(nomeItem)
	return fmt.Sprintf(
		"INSERT INTO %s (Nome)\n"+
			"VALUES ('%s')\n"+
			"ON CONFLICT(Nome) DO NOTHING;",
		itensTable,
		nome,
	)
}

// GenerateReciboCompraItemInsert gera o comando INSERT INTO para a tabela
// ReciboCompraItens a partir de um item extraído do PDF.
//
// codigoRecibo deve ser o mesmo valor gerado por NewCodigoRecibo e usado no
// INSERT de RecibosCompra correspondente. O ItemId é resolvido em tempo de
// execução do SQL via subquery que busca o Id na tabela Itens pelo Nome do
// item, então o INSERT do item (GenerateItemInsert) deve ser executado
// antes deste.
func GenerateReciboCompraItemInsert(codigoRecibo, nomeItem string, pesoKg, precoPorKg, total float64) string {
	nome := escapeSQLString(nomeItem)
	return fmt.Sprintf(
		"INSERT INTO %s (ReciboCodigo, ItemId, PesoKg, PrecoPorKg, Total)\n"+
			"VALUES (%s, (SELECT Id FROM %s WHERE Nome = '%s'), %s, %s, %s);",
		reciboCompraItensTable,
		escapeSQLString(codigoRecibo),
		itensTable,
		nome,
		formatFloat(pesoKg),
		formatFloat(precoPorKg),
		formatFloat(total),
	)
}

// formatFloat formata um float64 como literal numérico SQL, sem zeros à
// direita desnecessários (ex.: 394.53, 13.134).
func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}
