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
	i.Placeholder = "Type the word above and press Enter 👆🏽"
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
		input.Blink(m.textInput),
		spinner.Tick(m.spinner),
	}

	if m.timeMode {
		cmds = append(cmds, countDown())
	}

	return tea.Batch(cmds...)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		return m, nil

	case newWordMsg:
		m.currentWord = m.babbler.Babble()
		m.textInput.Reset()
		m.textInput.Focus()
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEnter:
			if m.textInput.Value() == m.currentWord {
				m.status = "🎉 correct!"
				m.points++
			} else {
				m.status = "😭 nope"
			}
			if !m.timeMode {
				return m, tea.Quit
			}
			return m, newWord()
		}
	}

	// TODO I think these 2 should also be handled within a a message
	var cmd tea.Cmd
	m.textInput, cmd = input.Update(msg, m.textInput)
	m.spinner, cmd = spinner.Update(msg, m.spinner)
	return m, cmd
}

func (m model) View() string {
	quitMsg := "(esc or ctrl-c to quit)"
	if m.done {
		return fmt.Sprintf("⏰ time is up! you had %d good answers\n%s", m.points, quitMsg)
	}

	s := termenv.
		String(spinner.View(m.spinner)).
		Foreground(term.Color("205")).
		String()

	timerText := fmt.Sprintf("%d seconds remaining\n%d points\n",
		int(m.timeLeft.Seconds()),
		m.points,
	)

	output := fmt.Sprintf(
		"%s\n%s\n%s\n\n%s",
		m.status,
		m.currentWord,
		input.View(m.textInput),
		quitMsg,
	)

	if m.timeMode {
		return fmt.Sprintf("%s %s\n%s\n",
			s,
			timerText,
			output,
		)
	}

	return fmt.Sprintf("%s%s\n", s, output)
}
