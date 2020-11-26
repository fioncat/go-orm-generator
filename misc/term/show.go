package term

import (
	"encoding/json"
	"fmt"
)

func Show(v interface{}) {
	data, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(data))

}
