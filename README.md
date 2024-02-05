# Program README

## Introduction
This program is designed to manage MySQL service remotely through SSH. It provides functionalities to check the current status of MySQL and restart it if needed. It is made in Go with `github.com/charmbracelet/bubbles/textinput` for interactive text input handling.
More details on why it is implemented [on my dev blog](https://byte-sized.fun/blog/restart-remote-mysql-server-with-go/)


## Features
- Check the current status of MySQL service remotely.
- Restart MySQL service remotely if necessary.

## Usage
1. Ensure that you have Go installed on your system.
2. Clone the repository or download the program files.
3. Navigate to the program directory.
4. Create a `.env` file with the required environment variables. Example:
```SERVER_URL=your_server_url```
5. Ensure that your SSH key (`id_ed25519`) is available in the `.ssh` directory within the program directory.
6. Build and run the program using the following command:
```go run .```
7. Follow the instructions provided by the program to check MySQL status or restart MySQL service if needed.
8. Type `yes` and press enter to restart MySQL service if it's already running.

**Note**: This program assumes that the MySQL PID file is located at `/var/run/mysqld/mysqld.pid` on the remote server. Please ensure that the actual PID file path matches this assumption. If your MySQL installation uses a different PID file location, you may need to modify the program accordingly.


## Disclaimer: Embedded SSH Key
Please note that this program embeds the SSH key (`id_ed25519`) within the built program. Sharing the built executable may expose this key, which poses a security risk. Exercise caution and avoid sharing the built program with untrusted parties. If distributing the program, consider providing instructions for users to set up their own SSH keys securely.

## Dependencies
- `github.com/joho/godotenv`: For loading environment variables from the `.env` file.
- `github.com/charmbracelet/bubbles/textinput`: For interactive text input handling.

## Note
- Make sure to provide the correct SSH credentials and server URL in the `.env` file for remote access.
- Ensure that the SSH private key (`id_ed25519`) is accessible to the program for establishing SSH connections.

## License
This program is distributed under the MIT License. See the `LICENSE` file for more information.
