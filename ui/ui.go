package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	authtypes "github.com/supabase-community/auth-go/types"

	"zel/lo/internal"
	"zel/lo/supabase"
)

// App states
const (
	viewAuth = iota
	viewMain
	viewCreateIssue
	viewListIssues
	viewMessage
)

// Styled components
var (
	accent = lipgloss.Color("212")
	errorColor = lipgloss.Color("204")
	surface = lipgloss.Color("236")
	muted = lipgloss.Color("246")

	appTitleStyle = lipgloss.NewStyle().Foreground(accent).Bold(true).Padding(0, 1)
	sectionTitleStyle = lipgloss.NewStyle().Foreground(accent).Bold(true)
	helpStyle = lipgloss.NewStyle().Foreground(muted)
	errorStyle = lipgloss.NewStyle().Foreground(errorColor)
	cardStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(accent).Padding(1, 2).Width(80)
)

// menu items
type menuItem struct{ title, desc string }

func (i menuItem) Title() string       { return i.title }
func (i menuItem) Description() string { return i.desc }
func (i menuItem) FilterValue() string { return i.title }

// Model

type Model struct {
	client   *supabase.Client
	userID   string
	view     int

	// Auth
	modeSignup bool
	emailInput    textinput.Model
	passwordInput textinput.Model
	nameInput     textinput.Model
	spinner       spinner.Model

	// Menu
	menu list.Model

	// Create issue
	titleInput       textinput.Model
	descriptionInput textarea.Model

	// List issues
	issues []internal.Issue

	// Message
	message string
	err     error
}

func New(client *supabase.Client, userID string) Model {
	email := textinput.New()
	email.Placeholder = "email@example.com"
	email.Width = 40
	email.Focus()
	password := textinput.New()
	password.EchoMode = textinput.EchoPassword
	password.Placeholder = "password"
	password.Width = 40
	name := textinput.New()
	name.Placeholder = "name (on signup)"
	name.Width = 40

	spin := spinner.New()
	spin.Spinner = spinner.Dot

	items := []list.Item{
		menuItem{"Create Issue", "Open a form to create a new issue"},
		menuItem{"List My Issues", "View issues you created"},
	}
	menu := list.New(items, list.NewDefaultDelegate(), 0, 0)
	menu.Title = "Menu"
	menu.SetShowHelp(false)
	menu.SetShowStatusBar(false)
	menu.SetShowPagination(false)

	title := textinput.New()
	title.Placeholder = "Issue title"
	title.CharLimit = 200
	title.Width = 50

	desc := textarea.New()
	desc.Placeholder = "Describe the issue..."
	desc.SetHeight(8)
	desc.SetWidth(76)

	initialView := viewAuth
	if userID != "" {
		initialView = viewMain
	}

	return Model{
		client:           client,
		userID:           userID,
		view:             initialView,
		modeSignup:       false,
		emailInput:       email,
		passwordInput:    password,
		nameInput:        name,
		spinner:          spin,
		menu:             menu,
		titleInput:       title,
		descriptionInput: desc,
	}
}

// tea.Model
func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.view {
		case viewAuth:
			return m.updateAuthKeys(msg)
		case viewMain:
			return m.updateMenuKeys(msg)
		case viewCreateIssue:
			return m.updateCreateKeys(msg)
		case viewListIssues:
			return m.updateListKeys(msg)
		case viewMessage:
			if key := msg.String(); key == "q" || key == "esc" || key == "enter" {
				m.view = viewMain
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		m.menu.SetSize(msg.Width-6, msg.Height-10)
	case messageErr:
		m.err = msg.error
		m.message = ""
		m.view = viewMessage
		return m, nil
	case messageInfo:
		m.err = nil
		m.message = msg.msg
		m.view = viewMessage
		return m, nil
	case changeView:
		m.view = msg.v
		if m.view == viewCreateIssue {
			m.titleInput.Focus()
			m.descriptionInput.Blur()
		}
		return m, nil
	case issuesMsg:
		m.issues = msg.list
		m.view = viewListIssues
		return m, nil
	}

	// Bubble updates
	switch m.view {
	case viewAuth:
		var cmds []tea.Cmd
		var cmd tea.Cmd
		m.emailInput, cmd = m.emailInput.Update(msg)
		cmds = append(cmds, cmd)
		m.passwordInput, cmd = m.passwordInput.Update(msg)
		cmds = append(cmds, cmd)
		m.nameInput, cmd = m.nameInput.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	case viewCreateIssue:
		var cmd tea.Cmd
		m.titleInput, cmd = m.titleInput.Update(msg)
		m.descriptionInput, _ = m.descriptionInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) View() string {
	switch m.view {
	case viewAuth:
		return m.viewAuth()
	case viewMain:
		return m.viewMenu()
	case viewCreateIssue:
		return m.viewCreateIssue()
	case viewListIssues:
		return m.viewListIssues()
	case viewMessage:
		return m.viewMessage()
	}
	return ""
}

// Auth
func (m *Model) updateAuthKeys(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch k.String() {
	case "tab":
		if m.modeSignup {
			if m.emailInput.Focused() {
				m.emailInput.Blur(); m.nameInput.Focus()
			} else if m.nameInput.Focused() {
				m.nameInput.Blur(); m.passwordInput.Focus()
			} else {
				m.passwordInput.Blur(); m.emailInput.Focus()
			}
		} else {
			if m.emailInput.Focused() {
				m.emailInput.Blur(); m.passwordInput.Focus()
			} else {
				m.passwordInput.Blur(); m.emailInput.Focus()
			}
		}
	case "ctrl+s":
		m.modeSignup = !m.modeSignup
		return m, nil
	case "enter":
		email := strings.TrimSpace(m.emailInput.Value())
		pass := m.passwordInput.Value()
		name := strings.TrimSpace(m.nameInput.Value())
		if email == "" || pass == "" || (m.modeSignup && name == "") {
			m.err = fmt.Errorf("fill required fields")
			return m, nil
		}
		if m.modeSignup {
			return m.submitSignup(email, pass, name)
		}
		return m.submitSignin(email, pass)
	case "esc", "q":
		return m, tea.Quit
	default:
		var cmds []tea.Cmd
		var cmd tea.Cmd
		m.emailInput, cmd = m.emailInput.Update(k)
		cmds = append(cmds, cmd)
		m.passwordInput, cmd = m.passwordInput.Update(k)
		cmds = append(cmds, cmd)
		if m.modeSignup {
			m.nameInput, cmd = m.nameInput.Update(k)
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

func (m *Model) submitSignin(email, pass string) (tea.Model, tea.Cmd) {
	m.err = nil
	return m, tea.Batch(func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
		defer cancel()
		_ = ctx
		session, err := m.client.SignInWithEmailPassword(email, pass)
		if err != nil {
			return messageErr{err}
		}
		m.userID = session.User.ID.String()
		return messageInfo{"Signed in"}
	}, func() tea.Msg { return changeView{viewMain} })
}

func (m *Model) submitSignup(email, pass, name string) (tea.Model, tea.Cmd) {
	m.err = nil
	return m, tea.Batch(func() tea.Msg {
		_, err := m.client.Auth.Signup(authtypes.SignupRequest{Email: email, Password: pass})
		if err != nil {
			return messageErr{err}
		}
		session, err := m.client.SignInWithEmailPassword(email, pass)
		if err != nil {
			return messageErr{err}
		}
		m.userID = session.User.ID.String()
		// Optional: create user profile record via main.CreateUser equivalent if needed
		_ = name
		return messageInfo{"Account created"}
	}, func() tea.Msg { return changeView{viewMain} })
}

// Menu
func (m *Model) updateMenuKeys(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch k.String() {
	case "enter":
		if it, ok := m.menu.SelectedItem().(menuItem); ok {
			switch it.title {
			case "Create Issue":
				m.view = viewCreateIssue
				m.titleInput.Focus()
				m.descriptionInput.Blur()
				return m, nil
			case "List My Issues":
				return m.fetchIssues()
			}
		}
	case "esc", "q":
		return m, tea.Quit
	}
	var cmd tea.Cmd
	m.menu, cmd = m.menu.Update(k)
	return m, cmd
}

// Create Issue
func (m *Model) updateCreateKeys(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch k.String() {
	case "esc":
		m.view = viewMain
		return m, nil
	case "tab":
		if m.titleInput.Focused() {
			m.titleInput.Blur(); m.descriptionInput.Focus()
		} else {
			m.descriptionInput.Blur(); m.titleInput.Focus()
		}
		return m, nil
	case "enter" :
		title := strings.TrimSpace(m.titleInput.Value())
		desc := strings.TrimSpace(m.descriptionInput.Value())
		if title == "" {
			m.err = fmt.Errorf("title required")
			return m, nil
		}
		return m.submitIssue(title, desc)
	default:
		var cmds []tea.Cmd
		var cmd tea.Cmd
		m.titleInput, cmd = m.titleInput.Update(k)
		cmds = append(cmds, cmd)
		m.descriptionInput, cmd = m.descriptionInput.Update(k)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}
	// no default fallthrough; all cases return
}

func (m *Model) submitIssue(title, desc string) (tea.Model, tea.Cmd) {
	return m, tea.Batch(func() tea.Msg {
		_, err := internal.CreateIssue(m.client, internal.CreateIssueRequest{Title: title, Description: desc}, m.userID)
		if err != nil {
			return messageErr{err}
		}
		return messageInfo{"Issue created"}
	}, func() tea.Msg { return changeView{viewMain} })
}

// List Issues
func (m *Model) updateListKeys(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch k.String() {
	case "esc", "q":
		m.view = viewMain
		return m, nil
	}
	return m, nil
}

func (m *Model) fetchIssues() (tea.Model, tea.Cmd) {
	return m, func() tea.Msg {
		issues, err := internal.ListIssues(m.client)
		if err != nil {
			return messageErr{err}
		}
		return issuesMsg{issues}
	}
}

// Messages
type (
	messageErr struct{ error }
	messageInfo struct{ msg string }
	changeView struct{ v int }
	issuesMsg struct{ list []internal.Issue }
)

func (m Model) viewAuth() string {
	mode := "Sign In"
	if m.modeSignup {
		mode = "Sign Up"
	}
	b := &strings.Builder{}
	fmt.Fprintln(b, appTitleStyle.Render("Zello"))
	fmt.Fprintln(b, cardStyle.Render(sectionTitleStyle.Render(mode)+
		"\n\nEmail:   "+m.emailInput.View()+
		"\nPassword: "+m.passwordInput.View()+
		func() string { if m.modeSignup { return "\nName:    "+m.nameInput.View() } ; return "" }()+
		"\n\nEnter to submit • Tab to switch • Ctrl+S toggle sign-in/sign-up • Q to quit"))
	if m.err != nil {
		fmt.Fprintln(b, errorStyle.Render(m.err.Error()))
	}
	return b.String()
}

func (m Model) viewMenu() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		appTitleStyle.Render("Zello"),
		cardStyle.Render(m.menu.View()),
		helpStyle.Render("Enter to select • Q to quit"),
	)
}

func (m Model) viewCreateIssue() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		sectionTitleStyle.Render("Create Issue"),
		cardStyle.Render("Title:\n"+m.titleInput.View()+"\n\nDescription:\n"+m.descriptionInput.View()+"\n\nEnter to submit • Tab to switch • Esc to back"),
	)
}

func (m Model) viewListIssues() string {
	if len(m.issues) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left,
			sectionTitleStyle.Render("My Issues"),
			cardStyle.Render("No issues found."+"\n\nEsc to back"),
		)
	}
	var b strings.Builder
	fmt.Fprintln(&b, sectionTitleStyle.Render("My Issues"))
	for _, is := range m.issues {
		line := fmt.Sprintf("#%-4d %-20s %-s", is.ID, is.Title, is.Description)
		fmt.Fprintln(&b, line)
	}
	return lipgloss.JoinVertical(lipgloss.Left, cardStyle.Render(b.String()+"\n\nEsc to back"))
}

func (m Model) viewMessage() string {
	msg := m.message
	if m.err != nil { msg = errorStyle.Render(m.err.Error()) }
	return lipgloss.JoinVertical(lipgloss.Left,
		sectionTitleStyle.Render("Info"),
		cardStyle.Render(msg+"\n\nEnter/Esc to continue"),
	)
}

func (m Model) UpdateMsg(msg tea.Msg) (Model, tea.Cmd) {
	switch v := msg.(type) {
	case messageErr:
		m.err = v.error
		m.message = ""
		m.view = viewMessage
	case messageInfo:
		m.err = nil
		m.message = v.msg
		m.view = viewMessage
	case changeView:
		m.view = v.v
	case issuesMsg:
		m.issues = v.list
		m.view = viewListIssues
	}
	return m, nil
}

// Program entry
func Start(client *supabase.Client, userID string) error {
	p := tea.NewProgram(New(client, userID))
	model, err := p.StartReturningModel()
	if err != nil { return err }
	_ = model
	return nil
}
