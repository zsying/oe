package main

import (
	"log"
	"os"

	"github.com/user/editor/internal/editor"
	"github.com/user/editor/internal/screen"
)

func main() {
	ed := editor.New()

	if len(os.Args) > 1 {
		if err := ed.OpenFile(os.Args[1]); err != nil {
			log.Fatalf("Failed to open file: %v", err)
		}
	}

	sc, err := screen.New(ed)
	if err != nil {
		log.Fatalf("Failed to initialize screen: %v", err)
	}

	if err := sc.Run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
