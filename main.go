package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	input "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/muesli/termenv"

	"github.com/Beartime234/babble"
)

const usage = `kboard [number] [time]

number: the number of words to generate. Must be a non-zero positive integer.
        defaults to 1 word.
time:   the number of seconds that the game will last.
        If none is passed, tha game finishes after the first word.

Examples:
 - kboard 2
 - kboard 1 30`

var term = termenv.ColorProfile()

func main() {
	if len(os.Args) < 2 || len(os.Args) > 3 {
		fmt.Println(usage)
		os.Exit(1)
	}

	if os.Args[1] == "-h" || os.Args[1] == "--help" {
		fmt.Println(usage)
		os.Exit(0)
	}

	numOfWords, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}

	if numOfWords < 1 {
		fmt.Println(usage)
		os.Exit(1)
	}

	var duration int
	if len(os.Args) == 3 {
		duration, err = strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatal(err)
			os.Exit(2)
		}

		if duration < 1 {
			fmt.Println(usage)
			os.Exit(1)
		}
	}

	p := tea.NewProgram(initialModel(numOfWords, duration))

	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}

type model struct {
	babbler     babble.Babbler
	textInput   input.Model
	spinner     spinner.Model
	currentWord string
	status      string
	startTime   time.Time
	duration    time.Duration
	timeLeft    time.Duration
	points      int
	done        bool
	timeMode    bool
}

func initialModel(numOfWords, duration int) model {
	b := babble.NewBabbler()
	b.Separator = " "
	b.Count = numOfWords

	i := input.NewModel()
	if numOfWords == 1 {
		i.Placeholder = "Type the word above and press Space or Enter 👆"
	} else {
		i.Placeholder = "Type the word above and press Enter 👆"
	}
	i.Focus()

	s := spinner.NewModel()
	s.Frames = spinner.Dot

	d := time.Duration(duration) * time.Second

	return model{
		babbler:     b,
		currentWord: b.Babble(),
		textInput:   i,
		spinner:     s,
		startTime:   time.Now(),
		duration:    d,
		timeLeft:    d,
		timeMode:    duration > 0,
	}
}

type newWordMsg struct{}

type countdownMsg struct {
	time time.Time
}

func newWord() tea.Cmd {
	return func() tea.Msg {
		return newWordMsg{}
	}
}

func countDown() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return countdownMsg{
			time: t,
		}
	})
}

func (m model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		input.Blink,
		spinner.Tick,
	}

	if m.timeMode {
		cmds = append(cmds, countDown())
	}

	return tea.Batch(cmds...)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var spinnerCmd tea.Cmd
	m.spinner, spinnerCmd = m.spinner.Update(msg)
	cmds := []tea.Cmd{spinnerCmd}

	switch msg := msg.(type) {
	case countdownMsg:
		if !m.timeMode {
			return m, nil
		}
		currentTime := msg.time
		timeConsumed := currentTime.Sub(m.startTime)
		if timeConsumed < m.duration {
			m.timeLeft = m.duration - timeConsumed
			return m, countDown()
		}
		m.done = true
		return m, tea.Quit

	case newWordMsg:
		m.currentWord = m.babbler.Babble()
		m.textInput.Reset()
		m.textInput.Focus()
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			return m, tea.Quit

		case "enter":
			return m.handleSubmission()

		case " ":
			if m.babbler.Count == 1 {
				return m.handleSubmission()
			}
			fallthrough

		default:
			var inputCmd tea.Cmd
			m.textInput, inputCmd = m.textInput.Update(msg)
			cmds = append(cmds, inputCmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	quitMsg := "(esc or ctrl-c to quit)"
	if m.done {
		var pointMsg string
		if m.points == 0 {
			pointMsg = "no good answers"
		} else if m.points == 1 {
			pointMsg = "1 good answer"
		} else {
			pointMsg = fmt.Sprintf("%d good answers", m.points)
		}
		return fmt.Sprintf("⏰ time is up! you had %s\n%s", pointMsg, quitMsg)
	}

	s := termenv.
		String(m.spinner.View()).
		Foreground(term.Color("205")).
		String()

	timerText := fmt.Sprintf("%d seconds remaining\n%d points\n",
		int(m.timeLeft.Seconds()),
		m.points,
	)

	if m.timeMode {
		timedOutput := fmt.Sprintf("%s%s\n  %s\n%s\n%s\n%s\n", s, timerText, m.currentWord, m.textInput.View(), m.status, quitMsg)
		return timedOutput
	}

	nonTimedOutput := fmt.Sprintf("%s%s\n%s\n%s\n%s\n", s, m.currentWord, m.textInput.View(), m.status, quitMsg)
	return nonTimedOutput
}

func (m model) handleSubmission() (model, tea.Cmd) {
	if m.textInput.Value() == m.currentWord {
		m.status = "🎉 correct!"
		m.points++
	} else {
		m.status = "😭 nope"
	}
	if !m.timeMode {
		m.textInput.Blur()
		return m, tea.Quit
	}
	return m, newWord()
}
