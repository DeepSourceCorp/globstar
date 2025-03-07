
import (
	"fmt"
	"os"
	// <expect-error>
	"unsafe"
)

func unsafeOperation() {
	var x int = 42
	ptr := unsafe.Pointer(&x) //unsafe operation
	fmt.Println("Unsafe operation with pointer:", ptr)
}


func safeOperation() {
	fmt.Println("Safe operation without unsafe package")
	fmt.Println("Current working directory:", os.Getwd())
}

func main() {
	fmt.Println("Demonstrating safe and unsafe imports in Go")
	safeOperation()
}
