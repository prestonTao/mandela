package utils

import (
	"fmt"
	"runtime"
)

func Go(f func()) {
	go func() {
		defer PrintPanicStack()
		f()
	}()
}

/*
	错误处理
*/
func PrintPanicStack() {
	if x := recover(); x != nil {
		fmt.Println(x)
		for i := 0; i < 10; i++ {
			funcName, file, line, ok := runtime.Caller(i)
			if ok {
				// fmt.Println("frame :[func:%s,file:%s,line:%d]\n", i, runtime.FuncForPC(funcName).Name(), file, line)
				Log.Error("%d frame :[func:%s,file:%s,line:%d]\n", i, runtime.FuncForPC(funcName).Name(), file, line)

				//				fmt.Println("frame " + strconv.Itoa(i) + ":[func:" + runtime.FuncForPC(funcName).Name() + ",file:" + file + ",line:" + line + "]\n")
			}
		}
	}
}
