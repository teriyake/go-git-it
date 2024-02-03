package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fogleman/ease"
	"github.com/lucasb-eyer/go-colorful"
	_ "github.com/muesli/reflow/indent"
	"github.com/muesli/reflow/padding"
	"github.com/muesli/termenv"
	"strconv"
	"strings"
	"time"
)

type (
	errMsg error
)

const (
	progressBarWidth  = 71
	progressFullChar  = "█"
	progressEmptyChar = "░"
)

var (
	appStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("235")).
			PaddingTop(2).
			PaddingLeft(4)
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Inherit(appStyle)
	menuTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("142"))
	docStyle       = lipgloss.NewStyle().Margin(1, 2)
	term           = termenv.EnvColorProfile()
	keyword        = makeFgStyle("142")
	subtle         = makeFgStyle("241")
	progressEmpty  = subtle(progressEmptyChar)
	dot            = colorFg(" • ", "236")
	ramp           = makeRamp("#B14FFF", "#00FFA3", progressBarWidth)
	focusedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("108"))
	blurredStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	focusedButton  = focusedStyle.Copy().Render("[ Submit ]")
	blurredButton  = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

func main() {
	fmt.Printf("\x1bc")
	initialModel := model{MainChoice: 0, SubChoice: 0, State: stateMain, Ticks: 60, Frames: 0, Progress: 0, Loaded: false, Quitting: false}
	p := tea.NewProgram(initialModel)
	if _, err := p.Run(); err != nil {
		fmt.Println("could not start program:", err)
	}
}

type tickMsg struct{}
type frameMsg struct{}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

func frame() tea.Cmd {
	return tea.Tick(time.Second/60, func(time.Time) tea.Msg {
		return frameMsg{}
	})
}

type state int

const (
	stateMain state = iota
	stateManageTask
	stateAddTask
)

type focus int

const (
	focusInput focus = iota
	focusTextarea
	focusButton
)

const (
	hotPink  = lipgloss.Color("#FF06B7")
	darkGray = lipgloss.Color("#767676")
)

var (
	inputStyle    = lipgloss.NewStyle().Foreground(hotPink)
	continueStyle = lipgloss.NewStyle().Foreground(darkGray)
)

type model struct {
	MainChoice int
	SubChoice  int
	State      state
	Ticks      int
	Frames     int
	Progress   float64
	Loaded     bool
	Quitting   bool
	Input      textinput.Model
	Textarea   textarea.Model
	Button     *string
	Focused    int
	Width      int
	Height     int
}

func (m model) Init() tea.Cmd {
	//tea.ClearScreen()
	tea.SetWindowTitle("ggi (go-git-it)")
	return tick()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width, m.Height = msg.Width, msg.Height
		return m, nil
	case tea.KeyMsg:
		switch m.State {
		case stateMain:
			return updateMain(msg, m)
		case stateManageTask:
			return updateManageTask(msg, m)
		case stateAddTask:
			return updateAddTask(msg, m)
		}
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			//tea.ClearScreen()
			return m, tea.Quit
		}

	case tickMsg:
		m.Ticks--
		if m.Ticks <= 0 {
			m.Quitting = true
			return m, tea.Quit
		}
		return m, tick()
	case frameMsg:
		m.Frames++
		m.Progress = ease.OutBounce(float64(m.Frames) / float64(100))
		if m.Progress >= 1 {
			m.Progress = 1
			m.Loaded = true
			m.Ticks = 3
			return m, tick()
		}
		return m, frame()
	}

	return m, nil
}

func (m model) View() string {
	ggiLogo := `
                     _|  
   _|_|_|    _|_|_|      
 _|    _|  _|    _|  _|  
 _|    _|  _|    _|  _|  
   _|_|_|    _|_|_|  _|  
       _|        _|      
   _|_|      _|_|        
	`

	if m.Quitting {
		return "\n  gg... i\n\n"
	}
	switch m.State {
	case stateMain:
		s := ggiLogo + "\n" + mainView(m) + "\n\n"
		return docStyle.Render(s)
		//return indent.String(s, 2)
	case stateAddTask:
		s := ggiLogo + "\n" + addTaskView(m) + "\n\n"
		return docStyle.Render(s)
		//return indent.String(s, 2)
	case stateManageTask:
		s := ggiLogo + "\n" + manageTaskView(m) + "\n\n"
		return docStyle.Render(s)
		//return indent.String(s, 2)
	default:
		return "Unknown state"
	}
}

func updateMain(msg tea.KeyMsg, m model) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "ctrl+c":
		m.Quitting = true
		return m, tea.Quit
	case "j", "down":
		m.MainChoice++
		if m.MainChoice > 4 {
			m.MainChoice = 4
		}
	case "k", "up":
		m.MainChoice--
		if m.MainChoice < 0 {
			m.MainChoice = 0
		}
	case "enter":
		switch m.MainChoice {
		case 1:
			m.State = stateAddTask
			m.Input = textinput.New()
			m.Input.Focus()
			m.Input.Placeholder = "  task name"
			m.Input.Prompt = "  "
			m.Textarea = textarea.New()
			m.Textarea.Placeholder = "task description"
			m.Textarea.Prompt = "| "
			m.Textarea.ShowLineNumbers = true
			m.Textarea.SetHeight(2)
			m.Button = &blurredButton
			m.Focused = 0

			return m, nil
		case 3:
			m.State = stateManageTask
			return m, nil
		default:
			return m, nil
		}
	default:
	}

	return m, nil
}

func updateManageTask(msg tea.KeyMsg, m model) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "ctrl+c":
		m.State = stateMain
	case "j", "down":
		m.SubChoice++
		if m.SubChoice > 2 {
			m.SubChoice = 2
		}
	case "k", "up":
		m.SubChoice--
		if m.SubChoice < 0 {
			m.SubChoice = 0
		}
	case "enter":
		// Implement action for sub menu selection
	}
	return m, nil
}

func mainView(m model) string {
	choices := []string{"Login", "Add a task", "Set up a list", "Manage a task", "About"}
	var b strings.Builder
	b.WriteString(colorBg("  Main Menu  ", "142") + "\n\n")
	for i, choice := range choices {
		if m.MainChoice == i {
			b.WriteString(fmt.Sprintf("> %s\n", keyword(choice)))
		} else {
			b.WriteString(fmt.Sprintf("  %s\n", choice))
		}
	}
	return padding.String((b.String() + "\n\n" + "Program quits in " + colorFg(strconv.Itoa(m.Ticks), "167") + " seconds" + "\n\n" + subtle("j/k, up/down: select") + dot + subtle("enter: choose") + dot + subtle("q, esc: quit")), 4)
}

func manageTaskView(m model) string {
	choices := []string{"Update deadline", "Update status", "Update description"}
	var b strings.Builder
	b.WriteString(colorBg("  Manage Task  ", "142") + "\n\n")
	for i, choice := range choices {
		if m.SubChoice == i {
			b.WriteString(fmt.Sprintf("> %s\n", keyword(choice)))
		} else {
			b.WriteString(fmt.Sprintf("  %s\n", choice))
		}
	}

	return padding.String((b.String() + "\n\n" + "Program quits in " + colorFg(strconv.Itoa(m.Ticks), "167") + " seconds" + "\n\n" + subtle("j/k, up/down: select") + dot + subtle("enter: choose") + dot + subtle("q, esc: quit")), 4)
}

func updateAddTask(msg tea.KeyMsg, m model) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd = make([]tea.Cmd, 2)

	switch msg.Type {
	case tea.KeyEnter:
		if m.Focused == 2 {
			// handle submit
			return m, tea.Quit
		}
	case tea.KeyCtrlC, tea.KeyEsc:
		return m, tea.Quit
	case tea.KeyShiftTab, tea.KeyCtrlP:
		m.prevInput()
	case tea.KeyTab, tea.KeyCtrlN:
		m.nextInput()
	}
	m.Input.Blur()
	m.Textarea.Blur()
	if m.Focused == 0 {
		m.Input.Focus()
		m.Textarea.Blur()
		m.Button = &blurredButton
	} else if m.Focused == 1 {
		m.Input.Blur()
		m.Textarea.Focus()
		m.Button = &blurredButton
	} else {
		m.Input.Blur()
		m.Textarea.Blur()
		m.Button = &focusedButton
	}

	m.Input, cmds[0] = m.Input.Update(msg)
	m.Textarea, cmds[1] = m.Textarea.Update(msg)
	return m, tea.Batch(cmds...)

}

func addTaskView(m model) string {
	var b strings.Builder
	b.WriteString(colorBg("  Add Task  ", "142") + "\n\n")
	b.WriteString(m.Input.View())
	b.WriteString("\n\n")
	b.WriteString(m.Textarea.View())
	b.WriteString("\n\n")
	b.WriteString(*m.Button)

	return padding.String((b.String() + "\n\n" + "Program quits in " + colorFg(strconv.Itoa(m.Ticks), "167") + " seconds" + "\n\n" + subtle("tab/shift+tab: focus") + dot + subtle("up/down: select line") + dot + subtle("enter: new line") + dot + subtle("esc: quit")), 4)
}

func (m *model) nextInput() {
	m.Focused = (m.Focused + 1) % 3
}

func (m *model) prevInput() {
	m.Focused--
	if m.Focused < 0 {
		m.Focused = 2
	}
}

func colorFg(val, color string) string {
	return termenv.String(val).Foreground(term.Color(color)).String()
}

func colorBg(val, color string) string {
	return termenv.String(val).Background(term.Color(color)).String()
}

func makeFgStyle(color string) func(string) string {
	return termenv.Style{}.Foreground(term.Color(color)).Styled
}

func makeRamp(colorA, colorB string, steps float64) (s []string) {
	cA, _ := colorful.Hex(colorA)
	cB, _ := colorful.Hex(colorB)

	for i := 0.0; i < steps; i++ {
		c := cA.BlendLuv(cB, i/steps)
		s = append(s, colorToHex(c))
	}
	return
}

func colorToHex(c colorful.Color) string {
	return fmt.Sprintf("#%s%s%s", colorFloatToHex(c.R), colorFloatToHex(c.G), colorFloatToHex(c.B))
}

func colorFloatToHex(f float64) (s string) {
	s = strconv.FormatInt(int64(f*255), 16)
	if len(s) == 1 {
		s = "0" + s
	}
	return
}
