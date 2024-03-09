package mongolog

import (
	"fmt"
	"os"
)

func ErrNotify(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
}
