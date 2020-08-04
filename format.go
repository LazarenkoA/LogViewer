package main

import (
	uuid "github.com/nu7hatch/gouuid"
	"strings"
)

type Iformatter interface {
	Format(string) map[string]string
}

type formatter1C struct{}

func (f *formatter1C) Format(str string) map[string]string {
	result := make(map[string]string, 0)
	tmp := make(map[string]string)

	// "," может встречаться в значение, в таких случаях значение будет строкой т.е. в "" или в ''
	for {
		if quoteStart := strings.Index(str, "'")+1; quoteStart > 0 {
			right := str[quoteStart:]
			if quoteEnd := strings.Index(right, "'"); quoteEnd >= 0 {
				if ID, err := uuid.NewV4(); err == nil {
					strID := ID.String()

					tmp[strID] = "'"+ str[quoteStart : quoteStart+quoteEnd] + "'"
					str = strings.Replace(str, tmp[strID], strID, -1)
				}
			} else {
				break
			}
		} else {
			break
		}
	}

	//r := csv.NewReader(strings.NewReader(str))
	//r.LazyQuotes = true
	//record, _ := r.Read()

	// проверяем на соответствие шаблону, важно при обработке многострочных логов
	// слишком дорагая операция, съедает 50% времени
	//re := regexp.MustCompile(`(?mi)\d\d:\d\d\.\d+[-]\d+`)
	//if ok := re.MatchString(str); !ok {
	//	return result
	//}

	parts := strings.Split(str, ",")
	if len(parts) < 2 {
		return result
	}

	// системные свойства, время, событие, длительность (06:11.062003-0,CLSTR,0,pro....)
	timeDuration := strings.Index(parts[0], "-")
	if timeDuration < 0 {
		return result
	}

	// время
	result["time"] = parts[0][:timeDuration]
	if timebreak := strings.Split(result["time"], "."); len(timebreak) > 0 {
		minsec := strings.Split(timebreak[0], ":")
		result["minutes"], result["seconds"] = minsec[0], minsec[1]
	}


	// длительность
	result["duration"] = parts[0][timeDuration+1:]

	// событие
	result["event"] = parts[1]

	for _, v := range parts {
		keyValue := strings.Split(strings.Trim(v, " "), "=")

		// могут быть такие данные
		// Descr='./src/ClusterDistribImpl.cpp(1640):60c686dc-798f-4d17-aadb-a90156a16eb8: Сеанс отсутствует или удаленID=30ecb789-2b56-46af-971d-c0a9579b9181
		if len(keyValue) >= 2 {
			result[keyValue[0]] = strings.Join(keyValue[1:], "=")
			for k, v := range tmp {
				result[keyValue[0]] = strings.Replace(result[keyValue[0]], k, v, -1)
			}

		}
	}

	return result
}
