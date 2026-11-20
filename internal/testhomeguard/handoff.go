package testhomeguard

import (
	"bufio"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const observerAddress = "CURATOR_TEST_HOME_GUARD_ADDRESS"
const observerKey = "CURATOR_TEST_HOME_GUARD_KEY"
const handoffTimeout = 5 * time.Second

// A fresh loopback capability works on Windows too and does not steal stdin or
// inherited fixture descriptors. A PID or nonce alone never admits a child:
// the live snapshot owner must authenticate this exact spawned PID/challenge.
// Capabilities and wire messages are never printed or written to diagnostics.
type liveObserver struct {
	*net.TCPListener
	key      string
	deadline time.Time
}

func newObserver() (*liveObserver, error) {
	l, err := net.ListenTCP("tcp4", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		return nil, err
	}
	deadline := time.Now().Add(handoffTimeout)
	_ = l.SetDeadline(deadline)
	return &liveObserver{TCPListener: l, key: randomChallenge(32), deadline: deadline}, nil
}

func randomChallenge(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic("testhomeguard: random source unavailable")
	}
	return hex.EncodeToString(b)
}

func proof(key, role, tuple string) string {
	m := hmac.New(sha256.New, []byte(key))
	_, _ = io.WriteString(m, role+"|"+tuple)
	return hex.EncodeToString(m.Sum(nil))
}

func readHandoff(c net.Conn) ([]string, error) {
	return readHandoffUntil(c, time.Now().Add(handoffTimeout))
}

func readHandoffUntil(c net.Conn, deadline time.Time) ([]string, error) {
	_ = c.SetDeadline(deadline)
	s, err := bufio.NewReader(io.LimitReader(c, 512)).ReadString('\n')
	if err != nil {
		return nil, err
	}
	return strings.Fields(s), nil
}

func (o *liveObserver) authenticate(childPID int) error {
	c, err := o.Accept()
	if err != nil {
		return err
	}
	defer func() { _ = c.Close() }()
	parts, err := readHandoffUntil(c, o.deadline)
	if err != nil || len(parts) != 3 || parts[0] != strconv.Itoa(childPID) || len(parts[1]) != 32 {
		return fmt.Errorf("invalid handoff")
	}
	tuple := strconv.Itoa(os.Getpid()) + "|" + parts[0] + "|" + parts[1]
	if !hmac.Equal([]byte(parts[2]), []byte(proof(o.key, "child", tuple))) {
		return fmt.Errorf("invalid handoff")
	}
	_, err = fmt.Fprintln(c, proof(o.key, "observer", tuple))
	return err
}

func verifyObserver() error {
	address, key := os.Getenv(observerAddress), os.Getenv(observerKey)
	host, port, err := net.SplitHostPort(address)
	p, parseErr := strconv.Atoi(port)
	if err != nil || parseErr != nil || host != "127.0.0.1" || p < 1 || p > 65535 || len(key) != 64 {
		return fmt.Errorf("absent handoff")
	}
	if _, err := hex.DecodeString(key); err != nil {
		return err
	}
	c, err := net.DialTimeout("tcp4", address, handoffTimeout) // #nosec G704 -- Literal 127.0.0.1 host and valid port required above; no DNS or remote endpoint.
	if err != nil {
		return err
	}
	defer func() { _ = c.Close() }()
	_ = c.SetDeadline(time.Now().Add(handoffTimeout))
	challenge := randomChallenge(16)
	tuple := strconv.Itoa(os.Getppid()) + "|" + strconv.Itoa(os.Getpid()) + "|" + challenge
	if _, err := fmt.Fprintln(c, os.Getpid(), challenge, proof(key, "child", tuple)); err != nil {
		return err
	}
	parts, err := readHandoff(c)
	if err != nil || len(parts) != 1 || !hmac.Equal([]byte(parts[0]), []byte(proof(key, "observer", tuple))) {
		return fmt.Errorf("unverified handoff")
	}
	return nil
}
