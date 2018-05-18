package log

import (
	"testing"
	"time"
)

func TestStclog_Trace(t *testing.T) {
	log := NewLog()
	log.Trace("wij","ssss")
	log.Trace("2222")
	log.Tracef("this is a test for %s","stclog")
	for i:=0 ;i<5 ;i++{
		go func(){
			log.Trace("sss",i)
		}()
	}
	time.Sleep(time.Second*10)
}
