package main

func recurseSelf(level int) {
	if level > 0 {
		recurseSelf(level - 2)
	}
}

func recurseOther(num uint) {
	recurseOrReturn(num - 1)
}

func recurseOrReturn(num uint) {
	if num == 0 {
		return
	}
	recurseOther(num - 1)
}
