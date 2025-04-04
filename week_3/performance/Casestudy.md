### Бенчмарк функции `SlowSearch`

Бюджет - нужно оптимизировать `SlowSearch` до значений `BenchmarkSolution-8 500 2782432 ns/op 559910 B/op 10422 allocs/op`
                                                                                1470783      1666125


```bash
goos: linux
goarch: amd64
pkg: hw3
cpu: 11th Gen Intel(R) Core(TM) i5-1135G7 @ 2.40GHz
BenchmarkSlow-8                2         816019716 ns/op        20229000 B/op     189823 allocs/op
BenchmarkFast-8                2         836604226 ns/op        20284340 B/op     189820 allocs/op
```

#### Пофайлинг памяти

```bash
   56.65MB 34.92% 34.92%    56.65MB 34.92%  regexp/syntax.(*compiler).inst (inline)
   28.30MB 17.44% 52.36%    28.30MB 17.44%  io.ReadAll
   20.51MB 12.64% 65.00%    20.51MB 12.64%  regexp/syntax.(*parser).newRegexp (inline)
   12.14MB  7.48% 72.48%   159.06MB 98.04%  hw3.SlowSearch
   10.26MB  6.33% 78.81%   103.04MB 63.51%  regexp.compile
    7.81MB  4.82% 83.63%    31.74MB 19.56%  regexp/syntax.parse
    3.20MB  1.97% 85.60%     3.20MB  1.97%  encoding/json.unquote (inline)
    2.93MB  1.81% 87.40%     5.86MB  3.61%  regexp/syntax.(*compiler).init (inline)
    2.91MB  1.79% 89.20%     2.91MB  1.79%  regexp.(*bitState).reset
    2.20MB  1.35% 90.55%     2.20MB  1.35%  reflect.mapassign_faststr0
```

Видно что основной точкой роста является то что весь файл грузится в пямять целиком, и вторая точка роста это компиляция регулярки на каждой итерации цикла

1. После того как сделал потоковую обработку и убрал повторную компиляцию регулярок получил такие результаты

```bash
goos: linux
goarch: amd64
pkg: hw3
cpu: 11th Gen Intel(R) Core(TM) i5-1135G7 @ 2.40GHz
BenchmarkSlow-8               55          20347595 ns/op        20194108 B/op     189835 allocs/op
BenchmarkFast-8              843           1470783 ns/op         1666125 B/op       8476 allocs/op
```
Видно что вышли на бюджет по ns/op и allocs/op но B/op пока больше в 3 раща

```bash
  43.35MB 21.35% 21.35%   120.80MB 59.48%  hw3.FastSearch
   43.03MB 21.19% 42.53%    43.03MB 21.19%  bufio.(*Scanner).Text (inline)
   33.48MB 16.48% 59.02%    33.48MB 16.48%  github.com/mailru/easyjson/jlexer.(*Lexer).String
   28.33MB 13.95% 72.97%    28.33MB 13.95%  regexp/syntax.(*compiler).inst (inline)
   13.80MB  6.80% 79.76%    13.80MB  6.80%  io.ReadAll
   10.26MB  5.05% 84.81%    10.26MB  5.05%  regexp/syntax.(*parser).newRegexp (inline)
    6.07MB  2.99% 87.80%    79.33MB 39.06%  hw3.SlowSearch
    5.14MB  2.53% 90.33%    51.53MB 25.37%  regexp.compile
    3.91MB  1.92% 92.26%    15.87MB  7.81%  regexp/syntax.parse
    1.65MB  0.81% 93.07%     1.65MB  0.81%  regexp.(*bitState).rese
```
Следующая точка роста bufio.(*Scanner).Text (inline)

2. Заменил `bufio.(*Scanner).Text (inline)` на работу с байтами `scanner.Bytes()`

```bash
goos: linux
goarch: amd64
pkg: hw3
cpu: 11th Gen Intel(R) Core(TM) i5-1135G7 @ 2.40GHz
BenchmarkSlow-8               52          20558489 ns/op        20176922 B/op     189830 allocs/op
BenchmarkFast-8              974           1086788 ns/op         1223501 B/op       5546 allocs/op
```
