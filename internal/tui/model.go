// Package tui implementa uma interface interativa de terminal (TUI) para
// visualizar um arquivo SQL gerado pelo pdftoinsert, navegando entre os
// comandos INSERT e o conteúdo SQL colorido.
package tui

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"golang.design/x/clipboard"

	"github.com/gstundner/pdftoinsert/internal/highlight"
	"github.com/gstundner/pdftoinsert/internal/sqlparser"
)

// focusArea identifica qual painel está atualmente com o foco do teclado.
type focusArea int

const (
	focusInserts focusArea = iota
	focusSQL
)

const (
	// leftPanelMinWidth e leftPanelMaxRatio controlam a largura do painel de
	// Inserts de forma responsiva ao tamanho do terminal.
	leftPanelMinWidth = 28
	leftPanelMaxRatio = 0.34

	// accentColor é a cor de destaque (roxo) usada em títulos, bordas em foco
	// e outros elementos de ênfase da interface.
	accentColor = lipgloss.Color("135")

	// horizontalScrollStep é o número de colunas movidas por pressionamento
	// das setas esquerda/direita no painel SQL.
	horizontalScrollStep = 8
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(accentColor)

	panelBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("240"))

	panelBorderFocusedStyle = panelBorderStyle.
				BorderForeground(accentColor)

	headerBorderStyle = panelBorderStyle

	lineNumberStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	selectedLineStyle = lipgloss.NewStyle().Background(lipgloss.Color("236"))

	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))

	copiedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42"))
)

// insertItem adapta sqlparser.Insert para a interface list.Item da bubbles.
type insertItem struct {
	insert sqlparser.Insert
}

func (i insertItem) Title() string { return i.insert.Summary }
func (i insertItem) Description() string {
	return fmt.Sprintf("linha %d · tabela %s", i.insert.Line, i.insert.Table)
}
func (i insertItem) FilterValue() string { return i.insert.Summary }

// Model é o modelo bubbletea da TUI de visualização de arquivos SQL.
type Model struct {
	filePath string
	file     *sqlparser.File

	highlightedLines []string
	// maxLineWidth é a largura (em colunas visíveis) da linha mais longa do
	// arquivo, usada para limitar a rolagem horizontal do painel SQL.
	maxLineWidth int

	insertsList list.Model
	sqlView     viewport.Model

	focus focusArea

	// xOffset é o deslocamento horizontal (em colunas) do painel SQL.
	xOffset int

	// copiedMessage é exibida brevemente na barra de ajuda após copiar um
	// INSERT para o clipboard.
	copiedMessage string

	width  int
	height int

	ready bool
	err   error
}

// New carrega o arquivo em path e constrói o modelo inicial da TUI.
func New(path string) (Model, error) {
	clipboard.Init()

	file, err := sqlparser.Load(path)
	if err != nil {
		return Model{}, err
	}

	highlighted, err := highlight.SQL(strings.Join(file.Lines, "\n"))
	if err != nil {
		return Model{}, err
	}

	maxLineWidth := 0
	for _, line := range file.Lines {
		if w := lipgloss.Width(line); w > maxLineWidth {
			maxLineWidth = w
		}
	}

	items := make([]list.Item, 0, len(file.Inserts))
	for _, ins := range file.Inserts {
		items = append(items, insertItem{insert: ins})
	}

	delegate := list.NewDefaultDelegate()
	insertsList := list.New(items, delegate, 0, 0)
	insertsList.Title = "[i] Inserts"
	insertsList.SetShowHelp(false)
	insertsList.SetShowStatusBar(false)
	insertsList.Styles.Title = titleStyle

	sqlView := viewport.New(0, 0)

	m := Model{
		filePath:         path,
		file:             file,
		highlightedLines: highlighted,
		maxLineWidth:     maxLineWidth,
		insertsList:      insertsList,
		sqlView:          sqlView,
		focus:            focusInserts,
	}
	m.refreshSQLView()

	return m, nil
}

// Init satisfaz tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update satisfaz tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.applySizes()
		m.ready = true
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "i":
			m.focus = focusInserts
			return m, nil
		case "s":
			m.focus = focusSQL
			return m, nil
		case "tab":
			if m.focus == focusInserts {
				m.focus = focusSQL
			} else {
				m.focus = focusInserts
			}
			return m, nil
		case "left":
			if m.focus == focusSQL {
				m.scrollHorizontal(-horizontalScrollStep)
			}
			return m, nil
		case "right":
			if m.focus == focusSQL {
				m.scrollHorizontal(horizontalScrollStep)
			}
			return m, nil
		case "y":
			if m.focus == focusInserts {
				m.copySelectedInsert()
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	switch m.focus {
	case focusInserts:
		prevIndex := m.insertsList.Index()
		m.insertsList, cmd = m.insertsList.Update(msg)
		if m.insertsList.Index() != prevIndex {
			m.copiedMessage = ""
			m.refreshSQLView()
		}
	case focusSQL:
		m.sqlView, cmd = m.sqlView.Update(msg)
	}

	return m, cmd
}

// scrollHorizontal ajusta o deslocamento horizontal do painel SQL em delta
// colunas, respeitando os limites do conteúdo.
func (m *Model) scrollHorizontal(delta int) {
	m.xOffset += delta
	if m.xOffset < 0 {
		m.xOffset = 0
	}

	maxOffset := m.maxLineWidth - m.sqlView.Width
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.xOffset > maxOffset {
		m.xOffset = maxOffset
	}

	m.refreshSQLView()
}

// currentInsert retorna o insertItem atualmente selecionado na lista de
// Inserts, se houver.
func (m Model) currentInsert() (sqlparser.Insert, bool) {
	item, ok := m.insertsList.SelectedItem().(insertItem)
	if !ok {
		return sqlparser.Insert{}, false
	}
	return item.insert, true
}

// copySelectedInsert copia o texto completo do comando INSERT atualmente
// selecionado no painel de Inserts para o clipboard.
func (m *Model) copySelectedInsert() {
	current, ok := m.currentInsert()
	if !ok {
		return
	}

	start := current.Line - 1

	// O fim do INSERT é a linha anterior ao início do próximo INSERT (ou o
	// final do arquivo, caso seja o último).
	end := len(m.file.Lines)
	for _, ins := range m.file.Inserts {
		if ins.Line > current.Line && ins.Line-1 < end {
			end = ins.Line - 1
		}
	}

	if start < 0 {
		start = 0
	}
	if end > len(m.file.Lines) {
		end = len(m.file.Lines)
	}

	lines := m.file.Lines[start:end]

	// Remove linhas em branco à direita (separadores entre INSERTs).
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}

	text := strings.Join(lines, "\n")

	if _, err := clipboard.Write(context.TODO(), clipboard.FmtText, []byte(text)); err != nil {
		m.copiedMessage = "falha ao copiar para o clipboard"
		return
	}

	m.copiedMessage = fmt.Sprintf("INSERT (linha %d) copiado para o clipboard!", current.Line)
}

// View satisfaz tea.Model.
func (m Model) View() string {
	if !m.ready {
		return "carregando..."
	}

	header := headerBorderStyle.
		Width(m.width - 2).
		Render(titleStyle.Render(filepath.Base(m.filePath)))

	leftBorder := panelBorderStyle
	sqlBorder := panelBorderStyle
	if m.focus == focusInserts {
		leftBorder = panelBorderFocusedStyle
	} else {
		sqlBorder = panelBorderFocusedStyle
	}

	sqlTitle := titleStyle.Render("[s] SQL")
	if pct := int(m.sqlView.ScrollPercent() * 100); len(m.highlightedLines) > 0 {
		sqlTitle = fmt.Sprintf("%s %s", sqlTitle, helpStyle.Render(fmt.Sprintf("(%d%%)", pct)))
	}

	leftPanel := leftBorder.Render(m.insertsList.View())

	rightPanel := sqlBorder.Render(
		lipgloss.JoinVertical(lipgloss.Left, sqlTitle, m.sqlView.View()),
	)

	body := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

	help := helpStyle.Render("i: foco Inserts · s: foco SQL · Tab: alternar · ↑/↓: navegar · ←/→: rolar SQL · y: copiar INSERT · q: sair")
	if m.copiedMessage != "" {
		help = copiedStyle.Render(m.copiedMessage)
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, body, help)
}

// applySizes recalcula as dimensões dos painéis com base no tamanho atual do
// terminal, mantendo o layout responsivo a qualquer largura/altura.
func (m *Model) applySizes() {
	const headerHeight = 3 // painel do nome do arquivo (borda + conteúdo + borda)
	const helpLines = 1
	const panelChromeHeight = 2 // bordas superior/inferior de cada painel
	const panelChromeWidth = 2  // bordas esquerda/direita de cada painel

	availableHeight := m.height - headerHeight - helpLines
	if availableHeight < 0 {
		availableHeight = 0
	}

	// Largura total do painel de Inserts (incluindo bordas), responsiva ao
	// tamanho do terminal: nunca menor que leftPanelMinWidth, nem maior que
	// leftPanelMaxRatio do total, respeitando o espaço disponível.
	leftTotalWidth := int(float64(m.width) * leftPanelMaxRatio)
	if leftTotalWidth < leftPanelMinWidth {
		leftTotalWidth = leftPanelMinWidth
	}
	if leftTotalWidth > m.width {
		leftTotalWidth = m.width
	}

	listHeight := availableHeight - panelChromeHeight
	if listHeight < 0 {
		listHeight = 0
	}
	listContentWidth := leftTotalWidth - panelChromeWidth
	if listContentWidth < 0 {
		listContentWidth = 0
	}
	m.insertsList.SetSize(listContentWidth, listHeight)

	sqlTotalWidth := m.width - leftTotalWidth
	sqlContentWidth := sqlTotalWidth - panelChromeWidth
	if sqlContentWidth < 0 {
		sqlContentWidth = 0
	}
	sqlHeight := availableHeight - panelChromeHeight - 1 // -1: linha de título dentro do painel
	if sqlHeight < 0 {
		sqlHeight = 0
	}
	m.sqlView.Width = sqlContentWidth
	m.sqlView.Height = sqlHeight

	m.refreshSQLView()
}

// refreshSQLView reconstrói o conteúdo do painel SQL, incluindo numeração de
// linhas, destaque da linha correspondente ao INSERT atualmente selecionado
// e recorte horizontal (xOffset) para permitir rolagem lateral, e rola a
// visualização até a linha apropriada.
func (m *Model) refreshSQLView() {
	if len(m.highlightedLines) == 0 {
		m.sqlView.SetContent("")
		return
	}

	selectedLine := 0
	if ins, ok := m.currentInsert(); ok {
		selectedLine = ins.Line
	}

	digits := len(fmt.Sprintf("%d", len(m.highlightedLines)))

	rendered := make([]string, len(m.highlightedLines))
	for i, code := range m.highlightedLines {
		lineNo := i + 1
		gutter := fmt.Sprintf("%*d │ ", digits, lineNo)
		gutterWidth := lipgloss.Width(gutter)

		// Recorta o código horizontalmente conforme o xOffset atual,
		// preservando os códigos ANSI de coloração (ansi.Cut é ciente de
		// sequências de escape e não as quebra).
		visibleWidth := m.sqlView.Width - gutterWidth
		if visibleWidth < 0 {
			visibleWidth = 0
		}
		code = ansi.Cut(code, m.xOffset, m.xOffset+visibleWidth)

		row := lineNumberStyle.Render(gutter) + code

		// Destaque da linha do INSERT selecionado na lista.
		if lineNo == selectedLine {
			row = selectedLineStyle.Render(row)
		}

		rendered[i] = row
	}

	m.sqlView.SetContent(strings.Join(rendered, "\n"))

	if selectedLine > 0 {
		target := selectedLine - m.sqlView.Height/2
		if target < 0 {
			target = 0
		}
		m.sqlView.SetYOffset(target)
	}
}
