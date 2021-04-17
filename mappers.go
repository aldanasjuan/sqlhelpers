package sqlhelpers

import "fmt"

func UpdateField(s string, i int) string {
	return fmt.Sprintf(`%v = $%d`, s, i+1)
}
