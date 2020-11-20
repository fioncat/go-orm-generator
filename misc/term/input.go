package term

import "fmt"

func Input(prompt string) string {
	fmt.Print(prompt + ": ")
	var input string
	fmt.Scanf("%s", &input)
	return input
}
