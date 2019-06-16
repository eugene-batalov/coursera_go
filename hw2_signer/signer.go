package main

import (
	"sort"
	"time"
	"sync"
	"fmt"
	"strings"
	"strconv"
)

// сюда писать код

func ExecutePipeline(hashSignJobs ... job) {
	inout := make([][]chan interface{}, len(hashSignJobs)+1)
	inout[0] = append(inout[0], make(chan interface{}, 100))
	j := 0
	for i, job := range hashSignJobs {
		switch i {
		case 0:
			inout[0] = append(inout[i], make(chan interface{}, 100))
			go job(inout[0][0], inout[0][1])
		OUT:
			for {
				select {
				case <-time.After(1 * time.Millisecond):
					break OUT
				case t := <-inout[0][1]:
					inout[1] = append(inout[1], make(chan interface{}, 100))
					inout[2] = append(inout[2], make(chan interface{}, 100))
					inout[1][j] <- t
					go hashSignJobs[1](inout[1][j], inout[2][j])
					j++
				}
			}
		case 1:
			break
		case 3:
			inout[3] = append(inout[i], make(chan interface{}, 100))
			inout[4] = append(inout[i], make(chan interface{}, 100))
			for n := 0; n < j; n++ {
				t := <-inout[2][n]
				inout[3][0] <- t
			}
			job(inout[3][0], inout[4][0])
		case 4:
			inout[5] = append(inout[i], make(chan interface{}, 100))
			job(inout[4][0], inout[5][0])
		default:
			for k := 0; k < j; k++ {
				inout[i] = append(inout[i], make(chan interface{}, 100))
				inout[i+1] = append(inout[i], make(chan interface{}, 100))
				go job(inout[i][k], inout[i+1][k])
			}
		}
	}
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
	v := <-in
	data := fmt.Sprintf("%v", v)
	crc32ch := make(chan string, 2)
	go fcrc32(data, crc32ch)
	md5 := callDataSignerMd5(data)
	go fcrc32(md5, crc32ch)
	crc32 := <-crc32ch
	result := crc32 + "~" + <-crc32ch
	out <- result
}

func MultiHash(in, out chan interface{}) {
	v := <-in
	data := fmt.Sprintf("%v", v)
	var result string
	var crc32ch []chan string
	for i := 0; i < 6; i++ {
		crc32ch = append(crc32ch, make(chan string, 6))
		go fcrc32(strconv.Itoa(i)+data, crc32ch[i])
	}
	var crc32slice []string
	for i := 0; i < 6; i++ {
		crc32slice = append(crc32slice, <-crc32ch[i])
	}
	result = strings.Join(crc32slice, "")
	out <- result
}

func CombineResults(in, out chan interface{}) {
	var result string
	var data []string
	for i := 0; i < 7; i++ {
		v := <-in
		data = append(data, v.(string))
	}
	sort.Strings(data)
	result = strings.Join(data, "_")
	out <- result
}
