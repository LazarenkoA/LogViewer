package main

import (
	"crypto/md5"
	"fmt"
	"testing"
)

const str = `00:03.869001-2998,CALL,2,process=rphost,p:processName=hrmcorp-n29,OSThread=8389,t:clientID=30190,t:applicationName=1CV8C,t:computerName=CA-TEST-RDP-1,t:connectID=331409,callWait=0,Usr=Парма,SessionID=533,Context=Система.ПолучитьИзВременногоХранилища,Interface=bc15bd01-10bf-413c-a856-ddc907fcd123,IName=IVResourceRemoteConnection,Method=0,CallID=25058,MName=send,Memory=-8208,MemoryPeak=589904,InBytes=0,OutBytes=0,CpuTime=1357 `

func BenchmarkGetHash(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getHashWitRace(str)
	}
}



func BenchmarkGetHashWithoutRace(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getHashWithoutRace(str)
	}
}

func getHashWithoutRace(s string) string {
	Sum := md5.Sum([]byte(s))
	return fmt.Sprintf("%x", Sum)
}

func getHashWitRace(inStr string) string {
	out := make(chan string)
	// race pattern
	sumString := func()  {
		Sum := md5.Sum([]byte(inStr))
		out <- fmt.Sprintf("%x", Sum)
	}

	// работа на опережедние
	go sumString()
	go sumString()
	go sumString()

	return <-out
}

// go test . -test.bench .*