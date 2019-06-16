package t

import (
	"log"
	"sort"
	"time"
	"sync"
	"fmt"
	"strings"
	"strconv"
)

// сюда писать код

func resendAndCloseOnTimeOut(in, out chan interface{}, waiter *sync.WaitGroup, i int) {
	defer waiter.Done()

	for {
		select {
		case <-time.After(10 * time.Millisecond):
			//close(in)
			log.Println("timer timed out, i=", i)
			return
		case result := <-in:
			//log.Printf("resending %v from in to out, i=%d\n", result, i)
			out <- result
		}
	}
}

func runPipeline(hashSignJobs [] job, in, out chan interface{}, outWg *sync.WaitGroup) {
	defer outWg.Done()
	var inout []chan interface{}
	inout = append(inout, in)
	inout = append(inout, make(chan interface{}, 100))
	wg := &sync.WaitGroup{}
	for i := range hashSignJobs {
		go hashSignJobs[i](inout[i*2], inout[i*2+1])
		inout = append(inout, make(chan interface{}, 100))
		inout = append(inout, make(chan interface{}, 100))
		wg.Add(1)
		chout := inout[i*2+2]
		if i == len(hashSignJobs)-1 {
			chout = out
		}
		go resendAndCloseOnTimeOut(inout[i*2+1], chout, wg, i)
	}
	wg.Wait()
}

func ExecutePipeline(hashSignJobs ... job) {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	var in []chan interface{}
	in = append(in, make(chan interface{}, 100))
	collect := make(chan interface{}, 100)
	out := make(chan interface{}, 100)
	wg := &sync.WaitGroup{}
	if len(hashSignJobs) < 3 {
		wg.Add(1)
		runPipeline(hashSignJobs, in[0], out, wg)
		return
	}
	num := 0
	hashSignJobs[0](in[0], in[0])
	last := len(hashSignJobs) - 1
OUT:
	for {
		select {
		case <-time.After(20 * time.Millisecond):
			break OUT
		case result := <-in[0]:
			num++
			in = append(in, make(chan interface{}, 100))
			in[num] <- result
			wg.Add(1)
			go runPipeline(hashSignJobs[1:last-1], in[num], collect, wg)
		}
	}
	wg.Wait()
	hashSignJobs[last-1](collect, out)
	hashSignJobs[last](out, out)
}

var (
	m      = &sync.Mutex{}
	fcrc32 = func(data string, ch chan string) {
		ch <- DataSignerCrc32(data)
	}
)

func callDataSignerMd5(data string) string {
	m.Lock()
	defer m.Unlock()
	return DataSignerMd5(data)
}

func SingleHash(in, out chan interface{}) {
OUT:
	for {
		select {
		case <-time.After(20 * time.Millisecond):
			break OUT
		case v := <-in:
			data := fmt.Sprintf("%v", v)
			log.Println("SingleHash got", data)
			crc32ch := make(chan string, 2)
			go fcrc32(data, crc32ch)
			md5 := callDataSignerMd5(data)
			go fcrc32(md5, crc32ch)
			crc32 := <-crc32ch
			log.Println("SingleHash crc32=", crc32)
			log.Println("SingleHash md5=", md5)
			result := crc32 + "~" + <-crc32ch
			log.Println("SingleHash sending out result", result)
			out <- result
		}
	}
}

func MultiHash(in, out chan interface{}) {
OUT:
	for {
		select {
		case <-time.After(1200 * time.Millisecond):
			break OUT
		case v := <-in:
			data := fmt.Sprintf("%v", v)
			log.Println("MultiHash got", data)
			var result string
			var crc32ch []chan string
			for i := 0; i < 6; i++ {
				crc32ch = append(crc32ch, make(chan string, 6))
				log.Println("MultiHash string(i)+data", strconv.Itoa(i)+data)
				go fcrc32(strconv.Itoa(i)+data, crc32ch[i])
			}
			var crc32slice []string
			for i := 0; i < 6; i++ {
				crc32slice = append(crc32slice, <-crc32ch[i])
			}
			log.Println("MultiHashes ", crc32slice)
			result = strings.Join(crc32slice, "")
			log.Println("MultiHash sending out", result)
			out <- result
		}
	}
}

func CombineResults(in, out chan interface{}) {
	var result string
	var data []string
OUT:
	for {
		select {
		case <-time.After(2900 * time.Millisecond):
			break OUT
		case v := <-in:
			data = append(data, fmt.Sprintf("%v", v))
		}
	}
	log.Println("CombineResults got", data)
	sort.Strings(data)
	app := "_"
	for i, str := range data {
		if i == len(data)-1 {
			app = ""
		}
		result += str + app
	}
	out <- result
}
