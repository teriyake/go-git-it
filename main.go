package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fogleman/ease"
	"github.com/lucasb-eyer/go-colorful"
	_ "github.com/muesli/reflow/indent"
	"github.com/muesli/reflow/padding"
	"github.com/muesli/reflow/wordwrap"
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
	maxWidth          = 64
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
	docStyle       = lipgloss.NewStyle().Margin(1, 2, 0)
	term           = termenv.EnvColorProfile()
	keyword        = makeFgStyle("142")
	subtle         = makeFgStyle("241")
	progressEmpty  = subtle(progressEmptyChar)
	dot            = colorFg(" • ", "236")
	ramp           = makeRamp("#B14FFF", "#00FFA3", progressBarWidth)
	focusedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("108"))
	blurredStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	focusedButton  = focusedStyle.Copy().Render("\t[ Confirm ]")
	blurredButton  = fmt.Sprintf("\t[ %s ]", blurredStyle.Render("Confirm"))
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	fmt.Printf("\x1bc")
	initialModel := model{MainChoice: 0, SubChoice: 0, State: stateMain, Ticks: 60, Frames: 0, Progress: 0, Loaded: false, Quitting: false, AltScreen: false}
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
	stateAbout
	stateSubmitting
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
	Spinner    spinner.Model
	Quitting   bool
	AltScreen  bool
	Input      textinput.Model
	Textarea   textarea.Model
	Inputs     []interface{}
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
		case stateAbout:
			return updateAbout(msg, m)
		case stateSubmitting:
			return updateSubmitting(msg, m)
		}
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			//tea.ClearScreen()
			m.Quitting = true
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
		fmt.Printf("\x1bc")
		return "\n\n  gg... i\n\n"
	}
	switch m.State {
	case stateMain:
		s := ggiLogo + "\n\n" + mainView(m) + "\n\n"
		return docStyle.Render(s)
		//return indent.String(s, 2)
	case stateAddTask:
		s := "\n\n\n" + addTaskView(m) + "\n\n"
		return docStyle.Render(s)
		//return indent.String(s, 2)
	case stateManageTask:
		s := ggiLogo + "\n\n" + manageTaskView(m) + "\n\n"
		return docStyle.Render(s)
		//return indent.String(s, 2)
	case stateAbout:
		s := ggiLogo + "\n\n" + wordwrap.String(aboutView(m), min(m.Width, maxWidth)) + "\n\n"
		return docStyle.Render(s)
	case stateSubmitting:
		s := "\n\n" + submittingView(m) + "\n\n"
		return docStyle.Render(s)
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
			m.Button = &blurredButton
			m.Focused = 0
			var inputs []interface{}
			title := textinput.New()
			title.Focus()
			title.Placeholder = "  task name"
			title.Prompt = "   "
			inputs = append(inputs, title)
			description := textarea.New()
			description.Placeholder = "task description"
			description.Prompt = "|"
			description.ShowLineNumbers = true
			description.SetHeight(8)
			inputs = append(inputs, description)
			dueDate := textinput.New()
			dueDate.Placeholder = "due on: YYYY-MM-DD"
			dueDate.Prompt = " "
			inputs = append(inputs, dueDate)
			m.Inputs = inputs

			m.AltScreen = !m.AltScreen
			return m, tea.EnterAltScreen
		case 3:
			m.State = stateManageTask
			m.AltScreen = !m.AltScreen
			return m, tea.EnterAltScreen
		case 4:
			m.State = stateAbout
			return m, nil
		default:
			return m, nil
		}
	default:
	}

	return m, nil
}

func mainView(m model) string {
	choices := []string{"Login", "Add a task", "Set up a to-do list", "Manage a task", "About"}
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

func updateManageTask(msg tea.KeyMsg, m model) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "ctrl+c":
		var cmd tea.Cmd
		if m.AltScreen {
			cmd = tea.ExitAltScreen
		} else {
			cmd = tea.Quit
			m.Quitting = true
		}
		m.AltScreen = !m.AltScreen
		return m, cmd
	case "shift+left":
		m.State = stateMain
		m.AltScreen = !m.AltScreen
		return m, tea.ExitAltScreen
	case "j", "down":
		m.SubChoice++
		if m.SubChoice > 4 {
			m.SubChoice = 4
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

func manageTaskView(m model) string {
	choices := []string{"Update status", "Update deadline", "Update description", "Collaborate", "Delete"}
	var b strings.Builder
	b.WriteString(colorBg("  Manage Task  ", "142") + "\n\n")
	for i, choice := range choices {
		if m.SubChoice == i {
			b.WriteString(fmt.Sprintf("> %s\n", keyword(choice)))
		} else {
			b.WriteString(fmt.Sprintf("  %s\n", choice))
		}
	}

	return padding.String((b.String() + "\n\n" + "Program quits in " + colorFg(strconv.Itoa(m.Ticks), "167") + " seconds" + "\n\n" + subtle("j/k, up/down: select") + dot + subtle("enter: choose") + dot + subtle("shift+left: back") + dot + subtle("q*2: quit")), 4)
}

func updateAbout(msg tea.KeyMsg, m model) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "ctrl+c":
		m.Quitting = true
		return m, tea.Quit
	case "shift+left":
		m.State = stateMain
		return m, nil
	}

	return m, nil
}

func aboutView(m model) string {
	var b strings.Builder
	b.WriteString(colorBg("  About  ", "142") + "\n\n")
	b.WriteString(fmt.Sprintf("This is the tui for ggi (go-git-it), a cli application that leverages Git functionalities to create and manage to-do tasks, update deadlines, and collaborate with friends on your agenda :)\n\nPlease visit the repo's homepage for more info: github.com/teriyake/go-git-it."))

	return padding.String((b.String() + "\n\n" + "Program quits in " + colorFg(strconv.Itoa(m.Ticks), "167") + " seconds" + "\n\n" + subtle("shift+left: back") + dot + subtle("q, esc: quit")), 4)

}

func updateAddTask(msg tea.KeyMsg, m model) (tea.Model, tea.Cmd) {

	switch msg.Type {
	case tea.KeyEnter:
		if m.Focused == len(m.Inputs) {
			// handle submit
			m.Spinner = spinner.New(spinner.WithSpinner(spinner.Line))
			m.State = stateSubmitting
			//m.AltScreen = !m.AltScreen
			return m, m.Spinner.Tick
		}
	case tea.KeyCtrlC, tea.KeyEsc:
		var cmd tea.Cmd
		if m.AltScreen {
			cmd = tea.ExitAltScreen
		} else {
			cmd = tea.Quit
			m.Quitting = true
		}
		m.AltScreen = !m.AltScreen
		return m, cmd
	case tea.KeyShiftLeft:
		m.State = stateMain
		m.AltScreen = !m.AltScreen
		return m, tea.ExitAltScreen
	case tea.KeyShiftTab, tea.KeyCtrlP:
		m.Focused--
		if m.Focused < 0 {
			m.Focused = len(m.Inputs) - 1
		}
		//m.prevInput()
	case tea.KeyTab, tea.KeyCtrlN:
		//m.nextInput()
		m.Focused = (m.Focused + 1) % (len(m.Inputs) + 1)
	}

	if m.Focused == len(m.Inputs) {
		m.Button = &focusedButton
	} else {
		m.Button = &blurredButton
	}

	for i, _ := range m.Inputs {
		if i != m.Focused {
			if textInput, ok := m.Inputs[i].(textinput.Model); !ok {
				if textArea, ok := m.Inputs[i].(textarea.Model); !ok {
					panic("unknown input type")
				} else {
					textArea.Blur()
					m.Inputs[i] = textArea
				}
			} else {
				textInput.Blur()
				m.Inputs[i] = textInput
			}
		} else {
			if textInput, ok := m.Inputs[i].(textinput.Model); !ok {
				if textArea, ok := m.Inputs[i].(textarea.Model); !ok {
					panic("unknown input type")
				} else {
					textArea.Focus()
					m.Inputs[i] = textArea
				}
			} else {
				textInput.Focus()
				m.Inputs[i] = textInput
			}
		}
	}

	var cmds []tea.Cmd = make([]tea.Cmd, len(m.Inputs))
	for i, input := range m.Inputs {
		if textInput, ok := input.(textinput.Model); !ok {
			if textArea, ok := input.(textarea.Model); !ok {
				panic("unknown input type")
			} else {
				textArea, cmds[i] = textArea.Update(msg)
				m.Inputs[i] = textArea
				continue
			}
		} else {
			textInput, cmds[i] = textInput.Update(msg)
			m.Inputs[i] = textInput
			continue
		}
	}

	return m, tea.Batch(cmds...)

}

func addTaskView(m model) string {
	var b strings.Builder
	b.WriteString(colorBg("  Add Task  ", "142") + "  " + "with an optional due date" + "\n\n")
	//b.WriteString(m.Input.View())
	b.WriteString(m.Inputs[0].(textinput.Model).View())
	b.WriteString("\n\n")
	//b.WriteString(m.Textarea.View())
	b.WriteString(m.Inputs[1].(textarea.Model).View())
	b.WriteString("\n\n")
	b.WriteString(m.Inputs[2].(textinput.Model).View())
	b.WriteString(*m.Button)

	return padding.String((b.String() + "\n\n" + "Program quits in " + colorFg(strconv.Itoa(m.Ticks), "167") + " seconds" + "\n\n" + subtle("tab/shift+tab: focus") + dot + subtle("up/down: select line") + dot + subtle("shift+left: back") + dot + subtle("esc*2: quit")), 4)
}

func updateSubmitting(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.Quitting = true
			return m, tea.Quit
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.Spinner, cmd = m.Spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func submittingView(m model) string {
	var b strings.Builder
	b.WriteString("Adding task... \n\n")

	return padding.String((m.Spinner.View() + "  " + b.String() + "\n\n" + "Program quits in " + colorFg(strconv.Itoa(m.Ticks), "167") + " seconds" + "\n\n" + subtle("shift+left: back") + dot + subtle("q, esc: quit")), 4)

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
