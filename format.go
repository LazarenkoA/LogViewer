package main

import (
	"fmt"
	"strconv"
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
	parseUnknown
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

type ParseError struct {
	code int
	what string
}

func (e ParseError) Error() string {
	return fmt.Sprintf("%v:%v", e.code, e.what)
}

func (f *formatter1C) Format(str string) (map[string]string, error) {
	if f.currentLine == nil {
		f.currentLine = make(map[string]string)
		f.parseValueState = parseValueBegin
	}

	for {
		if f.currentState == parseBegin {
			// Парсим начало строки лога
			durationBegin := strings.Index(str, "-")
			if durationBegin == -1 {
				return nil, ParseError{code: -1, what: "Expected '-'"}
			}

			durationEnd := strings.Index(str, ",")
			if durationEnd == -1 {
				return nil, ParseError{code: -1, what: "Expected ','"}
			}

			f.currentLine["time"] = str[:durationBegin]

			minutesIndex := strings.Index(f.currentLine["time"], ":")
			secondsIndex := strings.Index(f.currentLine["time"], ".")

			if minutesIndex == -1 || secondsIndex == -1 {
				return nil, ParseError{code: -1, what: "Minutes and seconds not found"}
			}

			minutes := f.currentLine["time"][:minutesIndex]
			minutesDig, err := strconv.Atoi(minutes)
			if err != nil {
				return nil, ParseError{code: -1, what: "Invalid minutes value"}
			} else if minutesDig >= 0 && minutesDig <= 59 {
				f.currentLine["minutes"] = minutes
			}

			seconds := f.currentLine["time"][(minutesIndex + 1):secondsIndex]
			secondsDig, err := strconv.Atoi(seconds)
			if err != nil {
				return nil, ParseError{code: -1, what: "Invalid seconds value"}
			} else if secondsDig >= 0 && secondsDig <= 59 {
				f.currentLine["seconds"] = seconds
			}

			duration := str[(durationBegin + 1):durationEnd]
			if _, err := strconv.Atoi(duration); err != nil {
				return nil, ParseError{code: -1, what: "Expected digits in duration"}
			}

			f.currentLine["duration"] = duration

			f.currentState = parseEvent
			str = str[(durationEnd + 1):]
		} else if f.currentState == parseEvent {
			index := strings.Index(str, ",")
			if index == -1 {
				return nil, ParseError{code: -1, what: "Expected ','"}
			}

			f.currentLine["event"] = str[:index]
			f.currentState = parseUnknown
			str = str[(index + 1):]
		} else if f.currentState == parseUnknown {
			index := strings.Index(str, ",")
			if index == -1 {
				return nil, ParseError{code: -1, what: "Expected ','"}
			}

			//f.currentLine["duration1"] = str[:index]
			f.currentState = parseKey
			str = str[(index + 1):]
		} else if f.currentState == parseKey {
			index := strings.Index(str, "=")
			if index == -1 {
				return nil, ParseError{code: -1, what: "Expected '='"}
			}

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
						return nil, ParseError{code: 0, what: "Not finished"}
					}

					var index = 0
					for {
						if index >= len(str) {
							return nil, ParseError{code: 0, what: "Not finished"} //еще не распарсили до конца
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
					// Может быть пустая строка, поэтому после цикла проверяем state
					for index, symbol := range str {
						if symbol == '\r' || symbol == '\n' || symbol == ',' {
							f.currentValue = str[:index]
							str = str[index:]
							f.parseValueState = parseValueFinish
							break
						}
					}

					// Если пустое значение или достигли конца строки
					if f.parseValueState == parseValueToNext {
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
					} else {
						return nil, ParseError{code: -1, what: "Invalid character at end of value"}
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

//////////////////////////////////
