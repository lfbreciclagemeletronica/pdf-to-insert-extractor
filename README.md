# pdftoinsert

CLI em Go que extrai dados de PDFs de resultado de pesagem e triagem (LFB) e
gera comandos SQL `INSERT` a partir deles.

## Sobre o projeto

O `pdftoinsert` lê um PDF no formato do relatório "RESULTADO DA PESAGEM E
TRIAGEM LFB" (há dois modelos de cabeçalho suportados: o modelo antigo, com
rótulos "PESO TOTAL"/"VALOR TOTAL", e o modelo novo, com rótulos
"PESO"/"VALOR" — ambos extraídos automaticamente pelo mesmo parser),
localiza os dados do cabeçalho (**Fornecedor**, **Peso Total**, **Valor
Total** e **Data**) e a tabela de materiais/itens, gerando um arquivo `.sql`
com os `INSERT`s correspondentes para as tabelas `Usuarios`,
`RecibosCompra`, `Itens` e `ReciboCompraItens`.

```sql
INSERT INTO Usuarios (Nome, Cpf, Cnpj, ChavePix)
VALUES ('Nome do Fornecedor', NULL, NULL, NULL)
ON CONFLICT(Nome) DO NOTHING;

INSERT INTO RecibosCompra (UsuarioId, DataRecibo, CodigoRecibo, ValorTotal, PesoTotalKg)
VALUES ((SELECT Id FROM Usuarios WHERE Nome = 'Nome do Fornecedor'), '15/05/2026', '28367', 394.53, 13.134);

INSERT INTO Itens (Nome)
VALUES ('Placa Notebook C')
ON CONFLICT(Nome) DO NOTHING;

INSERT INTO ReciboCompraItens (ReciboCodigo, ItemId, PesoKg, PrecoPorKg, Total)
VALUES (28367, (SELECT Id FROM Itens WHERE Nome = 'Placa Notebook C'), 0.262, 70, 18.34);

-- ... um par de INSERTs (Itens + ReciboCompraItens) para cada item da tabela de materiais do PDF
```

- `Cpf` e `Cnpj` não estão disponíveis no PDF e são gravados como `NULL`.
- `ChavePix` também é gravado como `NULL`, devendo ser preenchido manualmente depois.
- O `INSERT` de `Usuarios` (e o de `Itens`) usa `ON CONFLICT(Nome) DO NOTHING`
  para evitar duplicidade de forma nativa do SQLite (requer que `Nome` seja
  `UNIQUE` nas respectivas tabelas). A estrutura pode ser alterada para
  `DO UPDATE` no futuro se necessário para atualizar registros existentes.
- O `UsuarioId` em `RecibosCompra` e o `ItemId` em `ReciboCompraItens` são
  resolvidos em tempo de execução do SQL via subquery que busca o `Id` na
  tabela correspondente pelo Nome — por isso os `INSERT`s de `Usuarios` e
  `Itens` devem ser executados antes dos que dependem deles (o arquivo
  gerado já segue essa ordem).
- `DataRecibo` é gravada exatamente como aparece no PDF (formato `dd/mm/aaaa`).
- `CodigoRecibo` (em `RecibosCompra`) é um código aleatório de 5 dígitos
  (10000-99999), gerado uma única vez por PDF processado e reutilizado como
  `ReciboCodigo` em todos os `INSERT`s de `ReciboCompraItens` daquele recibo.
- `ValorTotal`, `PesoTotalKg`, `PesoKg`, `PrecoPorKg` e `Total` são
  convertidos do formato numérico brasileiro (vírgula decimal) para `REAL`.
- Os itens são extraídos da tabela de materiais do PDF (colunas
  Material/KG/Valor por KG/Total), localizada logo abaixo do cabeçalho.

Uso típico:

```powershell
pdftoinsert -f arquivo.pdf -o output.sql
```

## Dependências

- [Go](https://go.dev/dl/) 1.26 ou superior (necessário apenas para compilar; o binário final não requer Go instalado).
- [`spf13/cobra`](https://github.com/spf13/cobra) — parsing de comandos e flags da CLI.
- [`ledongthuc/pdf`](https://github.com/ledongthuc/pdf) — extração de texto plano de arquivos PDF.
- [`charmbracelet/bubbletea`](https://github.com/charmbracelet/bubbletea), [`bubbles`](https://github.com/charmbracelet/bubbles) e [`lipgloss`](https://github.com/charmbracelet/lipgloss) — interface interativa de terminal (`-v`/`--view`).
- [`alecthomas/chroma`](https://github.com/alecthomas/chroma) — coloração de sintaxe SQL na TUI.
- [`golang.design/x/clipboard`](https://github.com/golang-design/clipboard) — acesso cross-platform ao clipboard para copiar seleções do modo visual.

As dependências Go são gerenciadas via `go.mod`/`go.sum` e baixadas
automaticamente pelos comandos `go build`/`go run`/`go test`.

## Estrutura do projeto

- `main.go` — ponto de entrada.
- `cmd/` — definição dos comandos da CLI (cobra).
- `internal/extractor/` — extração de texto plano do PDF (ledongthuc/pdf).
- `internal/parser/` — extração dos campos do cabeçalho (Fornecedor, Peso Total, Valor Total, Data) e dos itens da tabela de materiais, a partir do texto, suportando os dois modelos de PDF.
- `internal/sqlgen/` — geração dos comandos SQL de INSERT (Usuarios, RecibosCompra, Itens, ReciboCompraItens).
- `internal/writer/` — escrita dos comandos SQL no arquivo de saída.
- `internal/sqlparser/` — localização dos comandos INSERT e tabelas em um arquivo SQL (usado pela TUI).
- `internal/highlight/` — coloração de sintaxe SQL via chroma (usado pela TUI).
- `internal/tui/` — interface interativa de terminal (`-v`/`--view`).
- `testdata/` — PDF de exemplo usado nos testes.
- `scripts/install.ps1` / `scripts/uninstall.ps1` — instalação/remoção do comando no PATH do sistema.

## Como rodar o projeto (desenvolvimento)

Compilar:

```powershell
go build -o pdftoinsert.exe .
```

Rodar sem compilar (via `go run`):

```powershell
go run . -f arquivo.pdf -o output.sql
```

Rodar os testes:

```powershell
go test ./...
```

## Instalando o comando no computador (uso pelo terminal)

Para poder rodar `pdftoinsert` diretamente de qualquer terminal (sem precisar
digitar `go run .` ou o caminho completo do binário), use o script de
instalação incluído no projeto:

```powershell
./scripts/install.ps1
```

O script:

1. Verifica se o Go está instalado.
2. Compila o binário `pdftoinsert.exe`.
3. Copia o binário para `%USERPROFILE%\.pdftoinsert\bin`.
4. Adiciona essa pasta ao `PATH` do usuário (caso ainda não esteja lá).

> Se o PATH for alterado, abra um **novo terminal** para que a mudança tenha efeito.

Depois disso, o comando fica disponível globalmente:

```powershell
pdftoinsert -f arquivo.pdf -o output.sql
```

Para desinstalar (remove o binário e a entrada do PATH):

```powershell
./scripts/uninstall.ps1
```

## Uso

### Extrair dados do PDF e gerar o SQL

```powershell
pdftoinsert -f arquivo.pdf -o output.sql
```

Flags:

- `-f`, `--file`: caminho do arquivo PDF de entrada.
- `-o`, `--output`: caminho do arquivo SQL de saída.

O programa localiza os campos do cabeçalho (Fornecedor, Peso Total, Valor
Total e Data) e os itens no texto do PDF e grava os `INSERT`s nas tabelas
`Usuarios`, `RecibosCompra`, `Itens` e `ReciboCompraItens` (ver seção "Sobre
o projeto" acima) no arquivo de saída.

Os `INSERT`s são **acrescentados ao final** do arquivo de saída (append),
sem apagar o conteúdo já existente — permitindo processar vários PDFs para o
mesmo arquivo `.sql`:

```powershell
pdftoinsert -f recibo1.pdf -o output.sql
pdftoinsert -f recibo2.pdf -o output.sql
# output.sql agora contém os INSERTs de ambos os recibos
```

### Visualizar um arquivo SQL gerado (TUI)

```powershell
pdftoinsert -v output.sql
```

Abre uma interface interativa de terminal (TUI) para revisar o arquivo SQL
gerado, com três painéis:

- **Cabeçalho** (topo): mostra o nome do arquivo em um painel com borda.
- **`[i] Inserts`** (esquerda): lista todos os comandos `INSERT` encontrados
  no arquivo, com a tabela afetada. Ao navegar com as setas, o painel `SQL`
  rola automaticamente até a linha correspondente.
- **`[s] SQL`** (direita): mostra o conteúdo completo do arquivo SQL com
  numeração de linhas e coloração de sintaxe (via [chroma](https://github.com/alecthomas/chroma)),
  com percentual de rolagem exibido no título. Suporta rolagem horizontal
  (`←`/`→`) para visualizar linhas mais longas que a largura do painel.

Atalhos de teclado:

- `i`: foca o painel Inserts.
- `s`: foca o painel SQL.
- `Tab`: alterna o foco entre os painéis.
- `↑`/`↓` (e `PgUp`/`PgDn`): navega verticalmente dentro do painel focado.
- `←`/`→` (no painel SQL): rola horizontalmente quando a linha ultrapassa a largura visível.
- `y` (no painel Inserts): copia o texto completo do INSERT selecionado para o clipboard.
- `q` ou `Ctrl+C`: fecha a TUI.

Implementada com o ecossistema [Charm](https://charm.sh/):
[Bubble Tea](https://github.com/charmbracelet/bubbletea),
[Bubbles](https://github.com/charmbracelet/bubbles) e
[Lip Gloss](https://github.com/charmbracelet/lipgloss).

## Roadmap

Atualmente são extraídos Fornecedor, Peso Total, Valor Total, Data e todos
os itens da tabela de materiais. O layout do PDF também contém CNPJ e
Inscrição Estadual que podem ser extraídos em extensões futuras seguindo o
mesmo padrão de `internal/parser`.
