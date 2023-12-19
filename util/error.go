package util

func PanicErr(err error) {
	if err == nil {
		return
	}
	// debug.PrintStack()
	panic(err)
}
