package main

import (
	"errors"
	"strings"
)

const (
	stateStart = iota
	stateDigital1
	stateDigital2
	stateDigital3
	stateDigital4
	stateDot
	stateDotDot
	stateDash
)

var (
	currentState int = 0
)

type Iformatter interface {
	Format(string) map[string]string
}

const (
	parseBegin = iota
	parseEvent
	parseDuration
	parseKey
	parseValue
	parseFinish
)

const (
	parseValueBegin = iota
	parseValueUntil
	parseValueToNext
	parseValueFinish
)

type formatter1C struct {
	currentState    int
	currentLine     map[string]string
	currentKey      string
	currentValue    string
	parseValueState int
	parseValueQuote uint8
}

func (f *formatter1C) Format(str string) (map[string]string, error) {
	// cat file.log
	// Txt='
	// fdaoijfdsjfiodsjf
	//'

	if f.currentLine == nil {
		f.currentLine = make(map[string]string)
		f.parseValueState = parseValueBegin
	}

	for {
		if f.currentState == parseBegin {
			delim := strings.Index(str, "-")
			index := strings.Index(str, ",")

			f.currentLine["time"] = str[:delim]

			minutes := strings.Index(f.currentLine["time"], ":")
			seconds := strings.Index(f.currentLine["time"], ".")

			f.currentLine["minutes"] = f.currentLine["time"][:minutes]
			f.currentLine["seconds"] = f.currentLine["time"][(minutes + 1):seconds]
			f.currentLine["duration"] = str[(delim + 1):index]

			f.currentState = parseEvent
			str = str[(index + 1):]
		} else if f.currentState == parseEvent {
			index := strings.Index(str, ",")
			f.currentLine["event"] = str[:index]

			f.currentState = parseDuration
			str = str[(index + 1):]
		} else if f.currentState == parseDuration {
			index := strings.Index(str, ",")
			f.currentLine["duration1"] = str[:index]

			f.currentState = parseKey
			str = str[(index + 1):]
		} else if f.currentState == parseKey {
			index := strings.Index(str, "=")
			f.currentKey = str[:index]

			f.currentState = parseValue
			f.parseValueState = parseValueBegin
			str = str[(index + 1):]
		} else if f.currentState == parseValue {
			for {
				if f.parseValueState == parseValueBegin {
					if len(str) == 0 || str[0] == ',' || str[0] == '\r' || str[0] == '\n' {
						// Key=,Key2=
						f.currentValue = ""
						f.parseValueState = parseValueFinish
					} else if str[0] == '\'' || str[0] == '"' {
						// Key="Value" || Key='Value'
						f.parseValueQuote = str[0]
						str = str[1:]
						f.parseValueState = parseValueUntil
					} else {
						// Key=Value
						f.parseValueState = parseValueToNext
					}
				} else if f.parseValueState == parseValueUntil {
					if len(str) == 0 {
						return make(map[string]string), errors.New("not finished")
					}

					var index = 0
					for {
						if index >= len(str) {
							return make(map[string]string), errors.New("not finished") //еще не распарсили до конца
						}

						if str[index] == '\'' || str[index] == '"' {
							if (index+1) < len(str) && str[index+1] == str[index] {
								index += 2
								continue
							} else if str[index] == f.parseValueQuote {
								f.currentValue = str[:index]
								str = str[(index + 1):]
								f.parseValueState = parseValueFinish
								break
							}
						}

						index += 1
					}
				} else if f.parseValueState == parseValueToNext {
					for index, symbol := range str {
						if symbol == '\r' || symbol == '\n' || symbol == ',' {
							f.currentValue = str[:index]
							str = str[index:]
							f.parseValueState = parseValueFinish
							break
						}
					}

					if f.parseValueState == parseValueToNext { // По сути это конец строки в принципе
						f.currentState = parseFinish
						f.currentLine[f.currentKey] = str
						break
					}
				} else if f.parseValueState == parseValueFinish {
					if len(str) == 0 {
						f.currentState = parseFinish
						f.currentLine[f.currentKey] = f.currentValue
					} else if str[0] == '\r' {
						f.currentState = parseFinish
						str = str[2:]
						f.currentLine[f.currentKey] = f.currentValue
					} else if str[0] == '\n' {
						f.currentState = parseFinish
						str = str[1:]
						f.currentLine[f.currentKey] = f.currentValue
					} else if str[0] == ',' {
						f.currentState = parseKey
						str = str[1:]
						f.currentLine[f.currentKey] = f.currentValue
					}
					break
				}
			}
		} else if f.currentState == parseFinish {
			ret := f.currentLine
			f.currentState = parseBegin
			f.currentKey = ""
			f.currentValue = ""
			f.currentLine = make(map[string]string)
			f.parseValueState = parseValueBegin
			f.parseValueQuote = 0
			return ret, nil
		}
	}
}

//////// Конечный автомат ////////

// Аналог шаблону регулярки \d\d:\d\d\.\d+[-]\d+
// Но быстрее
func FSM(str string) bool {
	defer func() { currentState = stateStart }()

	runes := []rune(str)
	for i := 0; i < len(runes); i++ {
		//fmt.Println(str[i:i+1])
		if isDigital(runes[i]) && (currentState == stateStart || currentState == stateDigital1) {
			currentState = stateDigital1
		} else if runes[i] == ':' && currentState == stateDigital1 {
			currentState = stateDotDot
		} else if isDigital(runes[i]) && (currentState == stateDotDot || currentState == stateDigital2) {
			currentState = stateDigital2
		} else if runes[i] == '.' && currentState == stateDigital2 {
			currentState = stateDot
		} else if isDigital(runes[i]) && (currentState == stateDot || currentState == stateDigital3) {
			currentState = stateDigital3
		} else if runes[i] == '-' && currentState == stateDigital3 {
			currentState = stateDash
		} else if isDigital(runes[i]) && (currentState == stateDash || currentState == stateDigital4) {
			currentState = stateDigital4
		} else if runes[i] == ',' && currentState == stateDigital4 {
			return true
		} else {
			return false
		}
	}

	return false
}

func isDigital(letter rune) bool {
	return letter >= '0' && letter <= '9'
}

//////////////////////////////////
