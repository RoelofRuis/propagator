benchmark:
	go test -bench=SolvePixelMatrix -cpuprofile=cpu.prof -memprofile=mem.prof -benchtime=60s ./examples/image

profile-cpu:
	go tool pprof cpu.prof

profile-mem:
	go tool pprof mem.prof