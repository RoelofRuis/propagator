benchmark:
	go test -bench=SolvePixelMatrix -cpuprofile=profile/cpu.prof -memprofile=profile/mem.prof -benchtime=60s ./examples/image

profile-cpu:
	go tool pprof profile/cpu.prof

profile-mem:
	go tool pprof profile/mem.prof