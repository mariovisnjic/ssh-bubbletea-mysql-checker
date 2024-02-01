package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"

	"github.com/joho/godotenv"
)

func main() {
	err := loadEnv()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	session, err := createSession()
	if err != nil {
		log.Fatalf("Failed to create SSH session: %v", err)
	}
	defer session.Close()

	mysqlStatus := checkMySQLStatus(session)
	if mysqlStatus == "true" {
		fmt.Println("MySQL pid found and running. No action required.")
	} else {
		fmt.Println("Cannot find MySQL. Trying to restart.")
		restartCommand := "service mysql restart"
		restartOutput, err := session.CombinedOutput(restartCommand)
		if err != nil {
			log.Fatalf("Failed to restart MySQL: %v", err)
		}
		fmt.Println("Restart MySQL output:", string(restartOutput))

		restartStatus := checkMySQLStatus(session)
		if restartStatus == "true" {
			fmt.Println("MySQL server was down but properly restarted.")
		} else {
			fmt.Println("MySQL server is down and could not be restarted. Please check manually.")
		}
	}
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

func checkMySQLStatus(session *ssh.Session) string {
	mysqlCheckCommand := `
		pidfile=/var/run/mysqld/mysqld.pid
		if [ -f $pidfile ]; then
			varpid=$(cat $pidfile)
			found=$(ps aux | awk '{print $2}' | grep -w $varpid)
			if [ -z "$found" ]; then
				echo "false"
			else
				echo "true"
			fi
		else
			echo "false"
		fi
	`

	output, err := session.CombinedOutput(mysqlCheckCommand)
	if err != nil {
		log.Fatalf("Failed to run MySQL check command: %v", err)
	}

	return strings.TrimSpace(string(output))
}
