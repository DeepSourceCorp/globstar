import (
	"fmt"

	// Insecure: Use of reflect package can lead to complex, unsafe, and less readable code.
	// <<expect-error> use of reflect package
	"reflect"
)

func main() {
	// Insecure Example: Using reflect to get the type of a variable
	var x = 42
	t := reflect.TypeOf(x)
	fmt.Printf("Insecure (reflect): Type of x is %s\n", t)

	// Why it's insecure or discouraged:
	// - Reflect code is harder to read and maintain.
	// - Can lead to runtime panics if used improperly.
	// - Often used to bypass type safety, increasing risk of bugs.

	// Secure Example: Use standard language features like fmt for type inspection
	y := "Hello, Go!"
	fmt.Printf("Secure: Value: %v, Type: %T\n", y, y)

	// Secure Example: Using type assertions or type switches instead of reflect
	checkType(y)
}

// Secure approach using type switches
func checkType(val interface{}) {
	switch v := val.(type) {
	case int:
		fmt.Println("Type is int:", v)
	case string:
		fmt.Println("Type is string:", v)
	default:
		fmt.Println("Unknown type")
	}
}
