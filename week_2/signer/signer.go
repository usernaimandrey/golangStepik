package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

func ExecutePipeline(workers ...job) {
	wg := &sync.WaitGroup{}
	in := make(chan interface{})
	workersCount := len(workers)
	wg.Add(workersCount)
	for _, worker := range workers {
		in = startWorker(worker, in, wg)
	}
	wg.Wait()

}

func startWorker(worker job, in chan interface{}, wg *sync.WaitGroup) chan interface{} {
	out := make(chan interface{})
	go func() {
		defer wg.Done()
		defer close(out)
		worker(in, out)
	}()
	return out
}

func SelectedType(i interface{}) (string, error) {
	switch i.(type) {
	case int:
		return fmt.Sprintf("%v", i.(int)), nil
	case string:
		return i.(string), nil
	default:
		fmt.Println("Unknow type")
		return "", fmt.Errorf("unknow type")
	}
}

func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}
	for inputData := range in {
		dataCast, err := SelectedType(inputData)

		if err != nil {
			panic("Data must be int or string")
		}
		wg.Add(1)
		go jobSingleHash(dataCast, out, wg, mu)
	}
	wg.Wait()
}

func jobSingleHash(inputData string, out chan interface{}, wg *sync.WaitGroup, mu *sync.Mutex) {
	defer wg.Done()
	var crcDataOne string
	var crcDataSecond string
	mu.Lock()
	md5Data := DataSignerMd5(inputData)
	mu.Unlock()

	crc32Chan := make(chan string)
	crc32Md5Chan := make(chan string)

	go func(data string, out chan string) {
		out <- DataSignerCrc32(data)
	}(inputData, crc32Chan)

	go func(md5Data string, out chan string) {
		out <- DataSignerCrc32(md5Data)
	}(md5Data, crc32Md5Chan)
	crcDataOne = <-crc32Chan
	crcDataSecond = <-crc32Md5Chan
	result := crcDataOne + "~" + crcDataSecond

	out <- result
}

func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for inputData := range in {
		dataCast, err := SelectedType(inputData)

		if err != nil {
			panic("Data must be int or string")
		}
		wg.Add(1)
		go jobMultiHash(dataCast, out, wg)
	}
	wg.Wait()
}

func jobMultiHash(data string, out chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	th := []int{0, 1, 2, 3, 4, 5}
	lengthTh := len(th)
	acc := make([]string, lengthTh, lengthTh)
	wgJob := &sync.WaitGroup{}
	mu := &sync.Mutex{}
	wgJob.Add(lengthTh)

	for i, val := range th {
		go func(val int, index int, wgJob *sync.WaitGroup, mu *sync.Mutex) {
			defer wgJob.Done()
			hash := DataSignerCrc32(strconv.Itoa(val) + data)
			mu.Lock()
			acc[index] = hash
			mu.Unlock()
		}(val, i, wgJob, mu)
	}
	wgJob.Wait()
	out <- strings.Join(acc, "")
}

func CombineResults(in, out chan interface{}) {
	acc := []string{}
	for inputData := range in {

		dataCast, err := SelectedType(inputData)
		if err != nil {
			panic("Data must be int or string")
		}
		acc = append(acc, dataCast)
	}

	sort.Strings(acc)

	out <- strings.Join(acc, "_")
}
