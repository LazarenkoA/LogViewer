package main

import (
	"fmt"
	"testing"
)

func Test_format1С(t *testing.T) {
	logs := []string{
`00:00.172000-1003,DBPOSTGRS,4,process=rphost,p:processName=hrm_temp,OSThread=20134,t:clientID=78722,t:applicationName=BackgroundJob,t:computerName=CA-T3-APP-1,t:connectID=80699,SessionID=1,Usr=DefUser,DBMS=DBPOSTGRS,DataBase=ca-t3-db-1\hrm_temp,Trans=0,dbpid=19753,Sql='SELECT Creation,Modified,Attributes,DataSize,BinaryData FROM Config WHERE FileName = $1 ORDER BY PartNo',Prm="p_1: 'a7fe0aa5-722c-4a09-889b-ee422421addc'::mvarchar",RowsAffected=1,Result=PGRES_TUPLES_OK`,
`00:06.558008-1003,TLOCK,4,process=rphost,p:processName=hrm_temp,OSThread=22516,t:clientID=84406,t:applicationName=BackgroundJob,t:computerName=CA-T3-APP-1,t:connectID=86289,SessionID=2,Usr=DefUser,DBMS=DBPOSTGRS,DataBase=ca-t3-db-1\hrm_temp,Regions=Reference65.REFLOCK,Locks='Reference65.REFLOCK Shared Fld1177=0',WaitConnections=,Context='ОбщийМодуль.ПолнотекстовыйПоискСервер.Модуль : 263 : ОбщегоНазначения.ПриНачалеВыполненияРегламентногоЗадания(Метаданные.РегламентныеЗадания.ОбновлениеИндексаППД);
        ОбщийМодуль.ОбщегоНазначения.Модуль : 5102 : Справочники.ВерсииРасширений.ЗарегистрироватьИспользованиеВерсииРасширений();
                Справочник.ВерсииРасширений.МодульМенеджера : 165 : ВерсияРасширений = ПараметрыСеанса.ВерсияРасширений;
                        МодульСеанса : 8 : СтандартныеПодсистемыСервер.УстановкаПараметровСеанса(ИменаПараметровСеанса);
                                ОбщийМодуль.СтандартныеПодсистемыСервер.Модуль : 52 : Справочники.ВерсииРасширений.УстановкаПараметровСеанса(ИменаПараметровСеанса, УстановленныеПараметры);
                                        Справочник.ВерсииРасширений.МодульМенеджера : 39 : ПараметрыСеанса.ВерсияРасширений = ВерсияРасширений();
                                                Справочник.ВерсииРасширений.МодульМенеджера : 719 : Блокировка.Заблокировать();'`,
`00:25.1850011000,TLOCK,4,process=rphost,p:processName=hrm_temp,OSThread=22517,t:clientID=84407`,
"жопа"}

	formatter := new(formatter1C)
	t.Parallel()
	for i, str := range logs {
		t.Run(fmt.Sprintf( "Строка %v", i+1), func(t *testing.T) {
			data := formatter.Format(str)

			// Обязательные поля
			if i != 2 && i != 3 {
				if _, ok := data["time"]; !ok {
					t.Error("Отсутствует свойство time")
				}
				if _, ok := data["event"]; !ok {
					t.Error("Отсутствует свойство event")
				}
				if v, ok := data["duration"]; ok {
					t.Error("Отсутствует свойство duration")
				} else if v != "1003" {
					t.Errorf("Некорректное свойство duration, ожидалось 1003, имеем %v", v)
				}
			}

			if i == 0 {
				if len(data) != 20 {
					t.Errorf("Некорретное разбиение, должно быть 20 частей, имеем %v", len(data))
				}
				if data["Sql"] != "'SELECT Creation,Modified,Attributes,DataSize,BinaryData FROM Config WHERE FileName = $1 ORDER BY PartNo'" {
					t.Error("Некорректно распарсилось свойство \"Sql\"")
				}
			}
			if i == 1 && len(data) != 18 {
				t.Errorf("Некорретное разбиение, должно быть 18 частей, имеем %v", len(data))
			}
			if (i == 2 || i == 3) && len(data) > 0 {
				t.Error("Строка не должна была распарситься")
			}
		})
	}
}

// go test -v -cover
// go test -coverprofile="cover.out"
// go tool cover -html="cover.out" -o cover.html