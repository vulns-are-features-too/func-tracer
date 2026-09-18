package main

func main() {
	start()

	runService(true)

	recurseSelf(2)
	recurseOther(5)

	extHello()
	extPrint("hi")

	nested0()

	_ = chain(true) || chain(false)

	end()
}

func start() {}

func end() {}
