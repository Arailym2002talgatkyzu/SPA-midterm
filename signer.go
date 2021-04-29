package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

// сюда писать код
const (
	hashNum = 6
)

//main step
var ExecutePipeline = func(jobs ...job) {
	wg := &sync.WaitGroup{}
	in := make(chan interface{})

	for _, job := range jobs {
		wg.Add(1)
		out := make(chan interface{})
		go startWorker(wg, in, out, job)
		in = out
	}
	wg.Wait()
}

//можно добавить и в ExecutePipeline
func startWorker(wg *sync.WaitGroup, in, out chan interface{}, j job) {
	defer wg.Done()
	j(in, out)
	close(out)
}

var SingleHash = func(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	mutex := &sync.Mutex{}
	for data := range in {
		wg.Add(1)
		go SHWorker(data, out, wg, mutex)
	}
	wg.Wait()
}

func SHWorker(data interface{}, out chan interface{}, wg *sync.WaitGroup, mutex *sync.Mutex) {
	crc32datachan := make(chan string)
	mutex.Lock()
	//calculating value of md5(data)

	md5data := DataSignerMd5(strconv.Itoa(data.(int)))
	mutex.Unlock()

	go func(out chan string, data interface{}) {
		//calculating value of crc32(data)
		crc32data := DataSignerCrc32(strconv.Itoa(data.(int)))
		out <- crc32data
	}(crc32datachan, data)

	crc32data := <-crc32datachan
	//calculating value of crc32(md5(data))
	crc32md5data := DataSignerCrc32(md5data)
	out <- crc32data + "~" + crc32md5data
	wg.Done()
}

var MultiHash = func(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for data := range in {
		wg.Add(1)
		go MHWorker(data, out, wg)
	}
	wg.Wait()
}

func MHWorker(data interface{}, out chan interface{}, wg *sync.WaitGroup) {
	mhwg := &sync.WaitGroup{}
	defer wg.Done()
	var mhdataresult = make([]string, hashNum, hashNum)
	for th := 0; th < hashNum; th++ {
		mhwg.Add(1)
		//calculating value of crc32(th+step1))
		crc32thstep1 := strconv.Itoa(th) + data.(string)
		go func(i int, wg *sync.WaitGroup, data string) {
			defer wg.Done()
			mhdataresult[i] = DataSignerCrc32(data)
		}(th, mhwg, crc32thstep1)
	}
	mhwg.Wait()
	result := strings.Join(mhdataresult, "")
	out <- result
}

//this was the easiest function, so I wrote it first
var CombineResults = func(in, out chan interface{}) {
	var allresults []string
	for data := range in {
		allresults = append(allresults, data.(string))
	}
	sort.Strings(allresults)
	result := strings.Join(allresults, "_")
	out <- result
}
