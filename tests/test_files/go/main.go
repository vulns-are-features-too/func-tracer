package main

func main() {
	start()

	svc := newService()
	svc.foo()
	svc.bar()

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
