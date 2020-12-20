package core

import (
	"fmt"
	"strings"

	"github.com/gookit/color"
)

func (t *GetInputWrapper) GetInputString(question string, def string) (text string, err error) {
	if def != "" {
		fmt.Printf("%s %s (press enter for default [%s]):", color.Yellow.Sprint("|=>"), question, def)
		text, err = t.Scanner.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("GetInputString %w", err)
		}

		text = strings.ReplaceAll(text, "\r\n", "")
		text = strings.ReplaceAll(text, "\n", "")

		if text == "" {
			text = def
		}
	} else {
		fmt.Printf("|=> %s:", question)
		text, err = t.Scanner.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("GetInputString %w", err)
		}
		text = strings.ReplaceAll(text, "\n", "")
	}

	return text, nil
}
