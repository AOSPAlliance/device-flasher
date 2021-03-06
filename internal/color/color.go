package color

import "fmt"

var (
	Blue   = Color("\033[1;34m%s\033[0m")
	Red    = Color("\033[1;31m%s\033[0m")
	Yellow = Color("\033[1;33m%s\033[0m")
)

func Color(color string) func(...interface{}) string {
	return func(args ...interface{}) string {
		return fmt.Sprintf(color, fmt.Sprint(args...))
	}
}
