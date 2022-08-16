package main

import (
	"crypto/md5"
	"fmt"
	"regexp"
	"testing"
)

const str = `00:03.869001-2998,CALL,2,process=rphost,p:processName=hrmcorp-n29,OSThread=8389,t:clientID=30190,t:applicationName=1CV8C,t:computerName=CA-TEST-RDP-1,t:connectID=331409,callWait=0,Usr=Парма,SessionID=533,Context=Система.ПолучитьИзВременногоХранилища,Interface=bc15bd01-10bf-413c-a856-ddc907fcd123,IName=IVResourceRemoteConnection,Method=0,CallID=25058,MName=send,Memory=-8208,MemoryPeak=589904,InBytes=0,OutBytes=0,CpuTime=1357 `

func BenchmarkRegexp(b *testing.B) {
	for i := 0; i < b.N; i++ {
		re := regexp.MustCompile(`(?mi)\d\d:\d\d\.\d+[-]\d+`)
		re.MatchString(str)
	}
}

func BenchmarkGetHashWithRace(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getHashWitRace(str)
	}
}

func BenchmarkGetHashWithoutRace(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getHashWithoutRace(str)
	}
}

func BenchmarkGetHashWitWorkers(b *testing.B) {
	//workersIn := make(chan string, 0)
	//workersOut := make(chan string, 0)
	//
	//for i := 0; i < b.N; i++ {
	//	workersIn <- str
	//}
	//
	//for i:=0; i<10; i++ {
	//	go func() {
	//		for s := range workersIn {
	//			Sum := md5.Sum([]byte(s))
	//			workersOut <- fmt.Sprintf("%x", Sum)
	//		}
	//	}()
	//}
	//
	//for range workersOut {
	//
	//}
}

func getHashWithoutRace(s string) string {
	Sum := md5.Sum([]byte(s))
	return fmt.Sprintf("%x", Sum)
}

func getHashWitRace(inStr string) string {
	out := make(chan string)
	// race pattern
	sumString := func() {
		Sum := md5.Sum([]byte(inStr))
		out <- fmt.Sprintf("%x", Sum)
	}

	// работа на опережедние
	go sumString()
	go sumString()
	go sumString()

	return <-out
}

func getHashWitWorkers(s string) string {

	Sum := md5.Sum([]byte(s))
	return fmt.Sprintf("%x", Sum)
}

// go test . -test.bench .*
