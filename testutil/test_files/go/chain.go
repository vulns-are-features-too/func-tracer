package main

func chain(bool) bool {
	return chain1() && chain2() || chain3()
}

func chain1() bool {
	return true
}

func chain2() bool {
	return false
}

func chain3() bool {
	return true
}
