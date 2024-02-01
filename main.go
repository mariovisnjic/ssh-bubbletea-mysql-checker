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
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	server := os.Getenv("SERVER_URL")
	port := 22
	user := "root"
	privateKeyPath := ".ssh/id_ed25519"

	// Read private key file
	key, err := os.ReadFile(privateKeyPath)
	if err != nil {
		log.Fatalf("Failed to load private key: %v", err)
	}

	// Parse private key
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("Failed to parse private key: %v", err)
	}

	// Create SSH config
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Accept any host key
	}

	// Connect to SSH server
	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", server, port), config)
	if err != nil {
		log.Fatalf("Failed to connect to SSH server: %v", err)
	}
	defer conn.Close()

	// Create new SSH session
	session, err := conn.NewSession()
	if err != nil {
		log.Fatalf("Failed to create SSH session: %v", err)
	}
	defer session.Close()

	// Define the command to check MySQL status
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

	// Run the MySQL check command remotely
	output, err := session.CombinedOutput(mysqlCheckCommand)
	if err != nil {
		log.Fatalf("Failed to run MySQL check command: %v", err)
	}

	// Parse the output to determine MySQL status
	mysqlStatus := strings.TrimSpace(string(output))
	if mysqlStatus == "true" {
		fmt.Println("MySQL pid found and running. No action required.")
	} else {
		// MySQL server not running, attempt to restart
		fmt.Println("Cannot find MySQL. Trying to restart.")
		restartCommand := "service mysql restart"
		restartOutput, err := session.CombinedOutput(restartCommand)
		if err != nil {
			log.Fatalf("Failed to restart MySQL: %v", err)
		}
		fmt.Println("Restart MySQL output:", string(restartOutput))

		// Check MySQL status again after restart
		restartOutput, err = session.CombinedOutput(mysqlCheckCommand)
		if err != nil {
			log.Fatalf("Failed to run MySQL check command after restart: %v", err)
		}
		restartStatus := strings.TrimSpace(string(restartOutput))
		if restartStatus == "true" {
			fmt.Println("MySQL server was down but properly restarted.")
		} else {
			fmt.Println("MySQL server is down and could not be restarted. Please check manually.")
		}
	}
}
