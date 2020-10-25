package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	input "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/Beartime234/babble"
)

const usage = `kboard [number]

number: the number of words to generate. Must be a non-zero positive integer.

Example: kboard 2`

func main() {
	if len(os.Args) != 2 {
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

	p := tea.NewProgram(initialModel(numOfWords))

	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}

type model struct {
	babbler     babble.Babbler
	textInput   input.Model
	currentWord string
}

func initialModel(numOfWords int) model {
	inputModel := input.NewModel()
	babbler := babble.NewBabbler()
	babbler.Separator = " "
	babbler.Count = numOfWords
	inputModel.Placeholder = "Type the word above and press Enter 👆🏽"
	inputModel.Focus()

	return model{
		babbler:     babbler,
		currentWord: babbler.Babble(),
		textInput:   inputModel,
	}
}

func (m model) Init() tea.Cmd {
	return input.Blink(m.textInput)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEnter:
			if m.textInput.Value() == m.currentWord {
				result := fmt.Sprintf("%s\n🎉 correct!", m.textInput.Value())
				m.textInput.SetValue(result)
				m.textInput.Blur()
			} else {
				result := fmt.Sprintf("%s\n😭 nope", m.textInput.Value())
				m.textInput.SetValue(result)
				m.textInput.Blur()
			}
			return m, tea.Quit
		}

	}

	m.textInput, cmd = input.Update(msg, m.textInput)
	return m, cmd
}

func (m model) View() string {
	return fmt.Sprintf(
		"%s\n\n\n%s\n\n%s",
		m.currentWord,
		input.View(m.textInput),
		"(esc or ctrl-c to quit)",
	) + "\n"
}
