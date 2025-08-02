package io

import (
	"fmt"
	"log"
	"strings"

	"github.com/mattn/go-tty"
)

var OverrideTTY string

func confirm(message string) bool {
	var t *tty.TTY
	var err error
	if OverrideTTY == "" {
		t, err = tty.Open()
	} else {
		t, err = tty.OpenDevice(OverrideTTY)
	}

	if err != nil {
		log.Println(err)
		log.Println("WARNING: cannot not confirm unless tty. should specify --force option to execute it.")
		return false
	}
	defer t.Close()

	for {
		_, err := fmt.Fprintf(t.Output(), "%s [y/n]: ", message)
		if err != nil {
			log.Fatal(err)
		}

		r, err := t.ReadString()
		if err != nil {
			log.Fatal(err)
		}

		switch strings.ToLower(strings.TrimSpace(r)) {
		case "y", "yes":
			return true
		case "n", "no":
			return false
		}
	}
}
