package term

import (
	"encoding/json"
	"fmt"
)

// Show outputs "v" in the form of formatted
// json in the terminal
func Show(v interface{}) {
	data, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(data))
}
