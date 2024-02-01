package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"

	"github.com/joho/godotenv"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	err := loadEnv()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type (
	errMsg error
)

type model struct {
	textInput   textinput.Model
	err         error
	mysqlStatus string
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "You don't have to write yes if it's already running"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	mysqlStatus := checkMySQLStatus()

	return model{
		textInput:   ti,
		err:         nil,
		mysqlStatus: mysqlStatus,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.textInput.Value() == "yes" {
				restartMySQL()
				m.mysqlStatus = checkMySQLStatus()
			}

		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}

	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return fmt.Sprintf(
		"Current MySQL status: %s\n\n%s\n\n%s\n\n%s",
		m.mysqlStatus,
		"Write yes and press enter if you want to restart MySQL:",
		m.textInput.View(),
		"(esc to quit)",
	)
}

func loadEnv() error {
	err := godotenv.Load()
	if err != nil {
		return err
	}
	return nil
}

func createSession() (*ssh.Session, error) {
	server := os.Getenv("SERVER_URL")
	port := 22
	user := "root"
	privateKeyPath := ".ssh/id_ed25519"

	key, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %v", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %v", err)
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", server, port), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SSH server: %v", err)
	}

	session, err := conn.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH session: %v", err)
	}

	return session, nil
}

func checkMySQLStatus() string {
	session, err := createSession()
	if err != nil {
		log.Fatalf("Failed to create SSH session: %v", err)
	}
	defer session.Close()

	mysqlCheckCommand := `
		pidfile=/var/run/mysqld/mysqld.pid
		if [ -f $pidfile ]; then
			varpid=$(cat $pidfile)
			found=$(ps aux | awk '{print $2}' | grep -w $varpid)
			if [ -z "$found" ]; then
				echo "Not running"
			else
				echo "Running"
			fi
		else
			echo "Not running"
		fi
	`

	output, err := session.CombinedOutput(mysqlCheckCommand)
	if err != nil {
		log.Fatalf("Failed to run MySQL check command: %v", err)
	}

	return strings.TrimSpace(string(output))
}

func restartMySQL() {
	session, err := createSession()
	if err != nil {
		log.Fatalf("Failed to create SSH session: %v", err)
	}
	defer session.Close()

	restartCommand := "service mysql restart"

	_, err = session.CombinedOutput(restartCommand)
	if err != nil {
		log.Fatalf("Failed to restart MySQL: %v", err)
	}
}
