package main

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

var _version = "1.0.3"

var (
	_defaultThreads = 100
	_maxThreads     = 50000
	_threads        = 100
	_apiKey         = "25febb1b5a003e935e16c0cd9662099a27ebec5e"
	_addrLen        = len("50679ea8C67095AF6bFF156f6dAFB82Dfc005Bfb")
)

var (
	success   int64
	failed    int64
	timeout   int64
	hostError int64

	rpcFailed int64

	traffics int64
)

func main() {
	println("testrpc version:", _version)
	args := os.Args
	if len(args) > 1 {
		threads, err := strconv.Atoi(args[1])
		if err != nil {
			println(fmt.Sprintf("threads error, use default %d threads, max threads is %d", _defaultThreads, _maxThreads))
		}
		if threads > 0 {
			_threads = int(math.Min(float64(threads), float64(_maxThreads)))
		} else {
			_threads = _defaultThreads
		}
	}
	if len(args) > 2 {
		apiKey := args[2]
		if apiKey != "" {
			_apiKey = apiKey
		}
	}

	url := ""
	if strings.HasPrefix(_apiKey, "http://") || strings.HasPrefix(_apiKey, "https://") {
		url = _apiKey
	} else {
		url = fmt.Sprintf("http://klaytn.testnet.blockpi.net/v1/rpc/%s", _apiKey)
	}

	println(fmt.Sprintf("Threads:\t%d \nAPI Endpoint:\t%s", _threads, url))

	reqFmt := "{\"jsonrpc\": \"2.0\",\"id\": \"1\",\"method\": \"eth_getBalance\",\"params\": [\"%s\",\"latest\"]}"
	for i := 0; i < _threads; i++ {
		go func() {
			for {
				req := fmt.Sprintf(reqFmt, randomAddr())
				PostJson(url, req)
				time.Sleep(time.Second * 1)
			}
		}()
	}

	for {
		time.Sleep(time.Second * 10)
		println(fmt.Sprintf("-------------------\ntime:\t%s\nsuccess:\t%d\nfailed:\t\t%d\ntimeout:\t%d\nhostError:\t%d\nrpcFailed:\t%d\ntraffics:\t%d", time.Now().String(), success, failed, timeout, hostError, rpcFailed, traffics))
	}
}

func PostJson(url string, jsonData string) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetContentType("application/json")
	req.Header.SetMethod("POST")

	req.SetRequestURI(url)

	req.SetBodyString(jsonData)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	client := fasthttp.Client{}

	if err := client.DoTimeout(req, resp, time.Second*10); err != nil {
		if err == fasthttp.ErrTimeout {
			atomic.AddInt64(&timeout, 1)
		} else if err == fasthttp.ErrNoFreeConns {
			atomic.AddInt64(&hostError, 1)
		} else {
			atomic.AddInt64(&hostError, 1)
			//println(err.Error())
		}
		atomic.AddInt64(&failed, 1)
		return
	}

	response := string(resp.Body())

	traffics += int64(len(jsonData) + len(response))

	if strings.Contains(response, "\"error\"") {
		atomic.AddInt64(&rpcFailed, 1)
		println(response)
	} else {
		atomic.AddInt64(&success, 1)
	}
}

func randomAddr() string {
	chars := "0123456789abcdef"
	addr := "0x"
	for i := 0; i < _addrLen; i++ {
		addr += string(chars[rand.Intn(16)])
	}
	return addr
}
