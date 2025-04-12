package golang

// <expect-error> Hidden goroutine
// func HiddenGoRoutine(){
// 	go func(){
// 		fmt.Println("Hidden goroutine")
// 	}()
// }

// func FunctionThatCallsGoRoutine(){
// 	fmt.Println("This is normal")
// 	go func(){
// 		fmt.Println("This is OK because FunctionThatCallsGoRoutine does other things")
// 	}()
// }

// // this also gets flagged
// func FunctionThatCallsGoroutineAlsoOk(){
// 	go func(){
// 		fmt.Println("This is OK because FunctionThatCallsGoroutineAlsoOk does other things")
// 	}
// 	fmt.Println("This is normal")
// }
