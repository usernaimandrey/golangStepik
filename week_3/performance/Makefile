report:
	go test -bench . -benchmem -cpuprofile=cpu.out -memprofile=mem.out -memprofilerate=1

open-mem:
	go tool pprof  hw3.test mem.out

open-cpu:
	go tool pprof  hw3.test cpu.out


test:
	go test -v

bench:
	go test -bench . -benchmem

