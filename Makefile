profile:
	go test -bench=Solve -cpuprofile=cpu.prof -benchtime=10s
	go tool pprof cpu.prof