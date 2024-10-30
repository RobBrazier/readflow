package prompt

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func YesNoPrompt(prompt string) (bool, error) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%s? [y/N] ", prompt)
		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		response = strings.ToLower(strings.TrimSpace(response))
		switch response {
		case "y", "yes":
			return true, nil
		case "n", "no", "":
			return false, nil
		}
	}
}

func TextPrompt(prompt string) (string, error) {
	fmt.Println(prompt)
	fmt.Print("> ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(response), nil
}
