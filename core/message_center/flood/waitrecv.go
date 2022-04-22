package flood

import (
	"mandela/config"
	"sync"
	"time"
	// "github.com/antlabs/timer"
)

const waitRequstTime = 30 //超时时间设置为60秒

var (
	waitRequest = new(sync.Map)
)

type HttpRequestWait struct {
	tagMap *sync.Map
}

/*
	等待请求返回
*/
func WaitRequest(class, tag string, timeout int64) (*[]byte, error) {
	if timeout <= 0 {
		timeout = waitRequstTime
	}

	// fmt.Println("1111111111111", class, tag)
	// engine.Log.Info("WaitRequest %s %s", class, tag)
	rwItr, ok := waitRequest.Load(class) //[class]
	if !ok {
		c := make(chan *[]byte, 1)
		hrw := HttpRequestWait{
			tagMap: new(sync.Map), //make(map[string]chan *[]byte),
		}
		hrw.tagMap.Store(tag, c)       //[tag] = c
		waitRequest.Store(class, &hrw) //[class] = &hrw
		ticker := time.NewTicker(time.Second * time.Duration(timeout))
		// timer.NewTimer().

		select {
		case <-ticker.C:
			hrw.tagMap.Delete(tag)
			return nil, config.ERROR_wait_msg_timeout
		case bs := <-c:
			ticker.Stop()
			return bs, nil
		}

	}
	rw := rwItr.(*HttpRequestWait)
	cItr, ok := rw.tagMap.Load(tag) // [tag]
	if !ok {
		c := make(chan *[]byte, 1)
		rw.tagMap.Store(tag, c) // [tag] = c

		ticker := time.NewTicker(time.Second * time.Duration(timeout))
		select {
		case <-ticker.C:
			rw.tagMap.Delete(tag)
			return nil, config.ERROR_wait_msg_timeout
		case bs := <-c:
			ticker.Stop()
			return bs, nil
		}
	}
	c := cItr.(chan *[]byte)

	ticker := time.NewTicker(time.Second * time.Duration(timeout))
	select {
	case <-ticker.C:
		rw.tagMap.Delete(tag)
		return nil, config.ERROR_wait_msg_timeout
	case bs := <-c:
		ticker.Stop()
		return bs, nil
	}
}

/*
	返回等待
*/
func ResponseWait(class, tag string, bs *[]byte) {
	// fmt.Println("ResponseWait", class, tag)
	// engine.Log.Info("ResponseWait %s %s", class, tag)
	rwItr, ok := waitRequest.Load(class) // [class]
	if !ok {
		return
	}
	rw := rwItr.(*HttpRequestWait)
	cItr, ok := rw.tagMap.Load(tag) // [tag]
	if !ok {
		return
	}
	c := cItr.(chan *[]byte)

	select {
	case c <- bs:
		return
	default:
	}
}
