Есть функиця, которая что-то там ищет по файлу. Но делает она это не очень быстро. Надо её оптимизировать.

Задание на работу с профайлером pprof.

Цель задания - научиться работать с pprof, находить горячие места в коде, уметь строить профиль потребления cpu и памяти, оптимизировать код с учетом этой информации. Написание самого быстрого решения не является целью задания.

```
$ go test -bench . -benchmem

goos: windows

goarch: amd64

BenchmarkSlow-8 10 142703250 ns/op 336887900 B/op 284175 allocs/op

BenchmarkSolution-8 500 2782432 ns/op 559910 B/op 10422 allocs/op

PASS
```

Запуск:
* `go test -v` - чтобы проверить что ничего не сломалось
* `go test -bench . -benchmem` - для просмотра производительности
* `go tool pprof -http=:8083 /path/ho/bin /path/to/out` - веб-интерфейс для pprof, пользуйтесь им для поиска горячих мест. Не забывайте, что у вас 2 режиме - cpu и mem, там разные out-файлы.

