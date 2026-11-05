package main

import (
	"crypto/tls"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/coder/websocket"
	"golang.org/x/crypto/ssh"
	"golang.org/x/text/encoding/ianaindex"
	"net/http"
)

func main() {
	fmt.Printf("cold repository ready: %T %T %T %T %T %T\n", http.Client{}, tls.Config{}, tea.NewProgram, websocket.Dial, ssh.NewClient, ianaindex.MIME)
}
