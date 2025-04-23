package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var errInvalidCredential = errors.New("invalid credential")

// PasswordDatabase loads from htpasswd-format text file. It only supported
// bcrypt hash.
type PasswordDatabase struct {
	lut map[string][]byte
}

func loadPasswordDatabase(path string) (*PasswordDatabase, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	lut := map[string][]byte{}
	scanner := bufio.NewScanner(f)
	for line := 1; scanner.Scan(); line++ {
		text := strings.TrimSpace(scanner.Text())
		if text == "" || text[0] == '#' {
			continue
		}
		username, passwordHash, found := strings.Cut(text, ":")
		if !found {
			return nil, fmt.Errorf("invalid entry at line %d", line)
		}
		lut[username] = []byte(passwordHash)
	}
	return &PasswordDatabase{lut}, nil
}

func (db *PasswordDatabase) Authenticate(username string, password []byte) error {
	passwordHash, ok := db.lut[username]
	if !ok {
		return errInvalidCredential
	}
	return bcrypt.CompareHashAndPassword(passwordHash, password)
}
