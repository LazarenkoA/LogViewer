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
"жопа",
`38:51.614002-1003,DBPOSTGRS,6,process=rphost,p:processName=acc-n20,OSThread=28531,t:clientID=2183,t:applicationName=WebServerExtension,t:computerName=CA-N20-WEB-1,t:connectID=25118,SessionID=31299,Usr=РысковаАА,AppID=WebClient,DBMS=DBPOSTGRS,DataBase=ca-n20-db-1\acc-n20,Trans=0,dbpid=4481,Sql="SELECTT1.AccountRRef,T1.Fld29328RRef,T1.Fld29331RRef,T1.Fld29335RRef,T1.Fld29336InitialBalanceDt_,T1.Fld29336InitialBalanceCt_,T1.Fld29336TurnoverDt_,T1.Fld29336TurnoverCt_,T1.Fld29336FinalBalanceDt_,T1.Fld29336FinalBalanceCt_,T1.Fld29331RRef,T7._Description,T7._Description,T8._EnumOrder,T1.AccountRRef,T9._Code,T9._Kind,T9._OrderFieldFROM (SELECTT2.Fld29328RRef AS Fld29328RRef,T2.AccountRRef AS AccountRRef,T2.Fld29331RRef AS Fld29331RRef,T2.Fld29335RRef AS Fld29335RRef,CASE WHEN SUM(T2.Fld29336TurnoverDt_) IS NULL THEN CAST(0 AS NUMERIC) ELSE SUM(T2.Fld29336TurnoverDt_) END AS Fld29336TurnoverDt_,CASE WHEN SUM(T2.Fld29336TurnoverCt_) IS NULL THEN CAST(0 AS NUMERIC) ELSE SUM(T2.Fld29336TurnoverCt_) END AS Fld29336TurnoverCt_,CASE WHEN SUM(T2.Fld29336Balance_) IS NULL THEN CAST(0 AS NUMERIC) WHEN MAX(T6._Kind) = CAST(0 AS NUMERIC) OR MAX(T6._Kind) = CAST(2 AS NUMERIC) AND SUM(T2.Fld29336Balance_) > CAST(0 AS NUMERIC) THEN SUM(T2.Fld29336Balance_) ELSE CAST(0 AS NUMERIC) END AS Fld29336InitialBalanceDt_,CASE WHEN SUM(T2.Fld29336Balance_) IS NULL THEN CAST(0 AS NUMERIC) WHEN MAX(T6._Kind) = CAST(1 AS NUMERIC) OR MAX(T6._Kind) = CAST(2 AS NUMERIC) AND SUM(T2.Fld29336Balance_) < CAST(0 AS NUMERIC) THEN -SUM(T2.Fld29336Balance_) ELSE CAST(0 AS NUMERIC) END AS Fld29336InitialBalanceCt_,CASE WHEN SUM(T2.Fld29336Balance_ + T2.Fld29336FinalTurnover_) IS NULL THEN CAST(0 AS NUMERIC) WHEN MAX(T6._Kind) = CAST(0 AS NUMERIC) OR MAX(T6._Kind) = CAST(2 AS NUMERIC) AND SUM(T2.Fld29336Balance_ + T2.Fld29336FinalTurnover_) > CAST(0 AS NUMERIC) THEN SUM(T2.Fld29336Balance_ + T2.Fld29336FinalTurnover_) ELSE CAST(0 AS NUMERIC) END AS Fld29336FinalBalanceDt_,CASE WHEN SUM(T2.Fld29336Balance_ + T2.Fld29336FinalTurnover_) IS NULL THEN CAST(0 AS NUMERIC) WHEN MAX(T6._Kind) = CAST(1 AS NUMERIC) OR MAX(T6._Kind) = CAST(2 AS NUMERIC) AND SUM(T2.Fld29336Balance_ + T2.Fld29336FinalTurnover_) < CAST(0 AS NUMERIC) THEN -SUM(T2.Fld29336Balance_ + T2.Fld29336FinalTurnover_) ELSE CAST(0 AS NUMERIC) END AS Fld29336FinalBalanceCt_,MAX(T6._Kind) AS AccKind_FROM (SELECTT3._Fld29328RRef AS Fld29328RRef,T3._AccountRRef AS AccountRRef,T3._Fld29331RRef AS Fld29331RRef,T3._Fld29335RRef AS Fld29335RRef,CASE WHEN T3._Period = '2020-01-01 00:00:00'::timestamp THEN T3._Fld29336 ELSE CAST(0 AS NUMERIC) END AS Fld29336Balance_,T3._Turnover29353 AS Fld29336FinalTurnover_,T3._TurnoverDt29351 AS Fld29336TurnoverDt_,T3._TurnoverCt29352 AS Fld29336TurnoverCt_FROM _AccRgAT029350 T3WHERE ((T3._Fld1265 = CAST(9039 AS NUMERIC))) AND (T3._Period >= '2020-01-01 00:00:00'::timestamp AND T3._Period < '2020-08-01 00:00:00'::timestamp AND ((T3._AccountRRef = '\\230\\313\\340\\313N\\300A\\367\\021\\345\\002\\224\\356\\007V\\342'::bytea)) AND ((T3._Fld29328RRef = '\\234~\\326\\2138\\021<\\221\\021\\351IQ\\247\\220\\367\\322'::bytea) AND (T3._Fld29328RRef = '\\234~\\326\\2138\\021<\\221\\021\\351IQ\\247\\220\\367\\322'::bytea)) AND (T3._Fld29336 <> CAST(0 AS NUMERIC) OR T3._TurnoverDt29351 <> CAST(0 AS NUMERIC) OR T3._TurnoverCt29352 <> CAST(0 AS NUMERIC) OR T3._Turnover29353 <> CAST(0 AS NUMERIC)))UNION ALL SELECTT4._Fld29328RRef AS Fld29328RRef,T4._AccountDtRRef AS AccountRRef,T4._Fld29331RRef AS Fld29331RRef,T4._Fld29335RRef AS Fld29335RRef,CAST(CAST(0 AS NUMERIC) AS NUMERIC(24, 2)) AS Fld29336Balance_,CAST(T4._Fld29336 AS NUMERIC(24, 2)) AS Fld29336FinalTurnover_,CAST(T4._Fld29336 AS NUMERIC(24, 2)) AS Fld29336TurnoverDt_,CAST(CAST(0 AS NUMERIC) AS NUMERIC(24, 2)) AS Fld29336TurnoverCt_FROM _AccRg29327 T4WHERE ((T4._Fld1265 = CAST(9039 AS NUMERIC))) AND (T4._Active = TRUE AND T4._AccountDtRRef <> '\\000\\000\\000\\000\\000\\000\\000\\000\\000\\000\\000\\000\\000\\000\\000\\000'::bytea AND ((T4._AccountDtRRef = '\\230\\313\\340\\313N\\300A\\367\\021\\345\\002\\224\\356\\007V\\342'::bytea)) AND ((T4._Fld29328RRef = '\\234~\\326\\2138\\021<\\221\\021\\351IQ\\247\\220\\367\\322'::bytea) AND (T4._Fld29328RRef = '\\234~\\326\\2138\\021<\\221\\021\\351IQ\\247\\220\\367\\322'::bytea)) AND (T4._Period > '2020-08-01 00:00:00'::timestamp OR T4._Period = '2020-08-01 00:00:00'::timestamp AND (T4._PeriodAdjustment >= CAST(0 AS NUMERIC))) AND (T4._Period < '2020-08-31 23:59:59'::timestamp OR T4._Period = '2020-08-31 23:59:59'::timestamp AND (T4._PeriodAdjustment <= CAST(0 AS NUMERIC))))UNION ALL SELECTT5._Fld29328RRef AS Fld29328RRef,T5._AccountCtRRef AS AccountRRef,T5._Fld29331RRef AS Fld29331RRef,T5._Fld29335RRef AS Fld29335RRef,CAST(CAST(0 AS NUMERIC) AS NUMERIC(24, 2)) AS Fld29336Balance_,CAST(-T5._Fld29336 AS NUMERIC(24, 2)) AS Fld29336FinalTurnover_,CAST(CAST(0 AS NUMERIC) AS NUMERIC(24, 2)) AS Fld29336TurnoverDt_,CAST(T5._Fld29336 AS NUMERIC(24, 2)) AS Fld29336TurnoverCt_FROM _AccRg29327 T5WHERE ((T5._Fld1265 = CAST(9039 AS NUMERIC))) AND (T5._Active = TRUE AND T5._AccountCtRRef <> '\\000\\000\\000\\000\\000\\000\\000\\000\\000\\000\\000\\000\\000\\000\\000\\000'::bytea AND ((T5._AccountCtRRef = '\\230\\313\\340\\313N\\300A\\367\\021\\345\\002\\224\\356\\007V\\342'::bytea)) AND ((T5._Fld29328RRef = '\\234~\\326\\2138\\021<\\221\\021\\351IQ\\247\\220\\367\\322'::bytea) AND (T5._Fld29328RRef = '\\234~\\326\\2138\\021<\\221\\021\\351IQ\\247\\220\\367\\322'::bytea)) AND (T5._Period > '2020-08-01 00:00:00'::timestamp OR T5._Period = '2020-08-01 00:00:00'::timestamp AND (T5._PeriodAdjustment >= CAST(0 AS NUMERIC))) AND (T5._Period < '2020-08-31 23:59:59'::timestamp OR T5._Period = '2020-08-31 23:59:59'::timestamp AND (T5._PeriodAdjustment <= CAST(0 AS NUMERIC))))) T2INNER JOIN _Acc36 T6ON T6._IDRRef = T2.AccountRRefWHERE (T6._Fld1265 = CAST(9039 AS NUMERIC))GROUP BY T2.Fld29328RRef,T2.AccountRRef,T2.Fld29331RRef,T2.Fld29335RRefHAVING (CASE WHEN SUM(T2.Fld29336TurnoverDt_) IS NULL THEN CAST(0 AS NUMERIC) ELSE SUM(T2.Fld29336TurnoverDt_) END) <> CAST(0 AS NUMERIC) OR (CASE WHEN SUM(T2.Fld29336TurnoverCt_) IS NULL THEN CAST(0 AS NUMERIC) ELSE SUM(T2.Fld29336TurnoverCt_) END) <> CAST(0 AS NUMERIC) OR (CASE WHEN SUM(T2.Fld29336Balance_) IS NULL THEN CAST(0 AS NUMERIC) WHEN MAX(T6._Kind) = CAST(0 AS NUMERIC) OR MAX(T6._Kind) = CAST(2 AS NUMERIC) AND SUM(T2.Fld29336Balance_) > CAST(0 AS NUMERIC) THEN SUM(T2.Fld29336Balance_) ELSE CAST(0 AS NUMERIC) END) <> CAST(0 AS NUMERIC) OR (CASE WHEN SUM(T2.Fld29336Balance_) IS NULL THEN CAST(0 AS NUMERIC) WHEN MAX(T6._Kind) = CAST(1 AS NUMERIC) OR MAX(T6._Kind) = CAST(2 AS NUMERIC) AND SUM(T2.Fld29336Balance_) < CAST(0 AS NUMERIC) THEN -SUM(T2.Fld29336Balance_) ELSE CAST(0 AS NUMERIC) END) <> CAST(0 AS NUMERIC) OR (CASE WHEN SUM(T2.Fld29336Balance_ + T2.Fld29336FinalTurnover_) IS NULL THEN CAST(0 AS NUMERIC) WHEN MAX(T6._Kind) = CAST(0 AS NUMERIC) OR MAX(T6._Kind) = CAST(2 AS NUMERIC) AND SUM(T2.Fld29336Balance_ + T2.Fld29336FinalTurnover_) > CAST(0 AS NUMERIC) THEN SUM(T2.Fld29336Balance_ + T2.Fld29336FinalTurnover_) ELSE CAST(0 AS NUMERIC) END) <> CAST(0 AS NUMERIC) OR (CASE WHEN SUM(T2.Fld29336Balance_ + T2.Fld29336FinalTurnover_) IS NULL THEN CAST(0 AS NUMERIC) WHEN MAX(T6._Kind) = CAST(1 AS NUMERIC) OR MAX(T6._Kind) = CAST(2 AS NUMERIC) AND SUM(T2.Fld29336Balance_ + T2.Fld29336FinalTurnover_) < CAST(0 AS NUMERIC) THEN -SUM(T2.Fld29336Balance_ + T2.Fld29336FinalTurnover_) ELSE CAST(0 AS NUMERIC) END) <> CAST(0 AS NUMERIC)) T1LEFT OUTER JOIN _Reference138 T7ON (T1.Fld29331RRef = T7._IDRRef) AND (T7._Fld1265 = CAST(9039 AS NUMERIC))LEFT OUTER JOIN _Enum886 T8ON T1.Fld29335RRef = T8._IDRRefLEFT OUTER JOIN _Acc36 T9ON (T1.AccountRRef = T9._IDRRef) AND (T9._Fld1265 = CAST(9039 AS NUMERIC))WHERE (T1.Fld29328RRef = '\\234~\\326\\2138\\021<\\221\\021\\351IQ\\247\\220\\367\\322'::bytea) AND (T1.Fld29328RRef = '\\234~\\326\\2138\\021<\\221\\021\\351IQ\\247\\220\\367\\322'::bytea)",RowsAffected=18,Result=PGRES_TUPLES_OK,Context='Форма.Вызов : Отчет.ОборотноСальдоваяВедомостьПоСчету.Форма.ФормаОтчета.Модуль.СформироватьОтчетСерверОтчет.ОборотноСальдоваяВедомостьПоСчету.Форма.ФормаОтчета.Форма : 560 : ОбъектОтчет.СформироватьОтчет(Результат, ДанныеРасшифровки, СхемаКомпоновкиДанных, Истина); Отчет.ОборотноСальдоваяВедомостьПоСчету.МодульОбъекта : 343 : СтандартныеОтчеты.ВывестиОтчет(ЭтотОбъект, Результат, ДанныеРасшифровки, Схема, ВыводитьПолностью); ОбщийМодуль.СтандартныеОтчеты.Модуль : 612 : ПроцессорВывода.Вывести(ПроцессорКомпоновки, Истина);'`,
`100:235.185001-1000,TLOCK,4,process=rphost,p:processName=hrm_temp,OSThread=22517,t:clientID=84407`,
`00:06.558008-1003,TLOCK,4,process=rphost,p:processName=hrm_temp,OSThread=22516,t:clientID=84406,t:applicationName=BackgroundJob,t:computerName=CA-T3-APP-1,t:connectID=86289,SessionID=2,Usr=DefUser,DBMS=DBPOSTGRS,DataBase=ca-t3-db-1\hrm_temp,Regions=Reference65.REFLOCK,Locks='Reference65.REFLOCK Shared Fld1177=0',WaitConnections=,Context="ОбщийМодуль.ПолнотекстовыйПоискСервер.Модуль : 263 : ОбщегоНазначения.ПриНачалеВыполненияРегламентногоЗадания(Метаданные.РегламентныеЗадания.ОбновлениеИндексаППД);
        ОбщийМодуль.ОбщегоНазначения.Модуль : 5102 : Справочники.ВерсииРасширений.ЗарегистрироватьИспользованиеВерсииРасширений();
                Справочник.ВерсииРасширений.МодульМенеджера : 165 : ВерсияРасширений = ПараметрыСеанса.ВерсияРасширений;
                        МодульСеанса : 8 : СтандартныеПодсистемыСервер.УстановкаПараметровСеанса(ИменаПараметровСеанса);
                                ОбщийМодуль.СтандартныеПодсистемыСервер.Модуль : 52 : Справочники.ВерсииРасширений.УстановкаПараметровСеанса(ИменаПараметровСеанса, УстановленныеПараметры);
                                        Справочник.ВерсииРасширений.МодульМенеджера : 39 : ПараметрыСеанса.ВерсияРасширений = ВерсияРасширений();
                                                Справочник.ВерсииРасширений.МодульМенеджера : 719 : Блокировка.Заблокировать();"`}

	formatter := new(formatter1C)
	t.Parallel()
	for i, str := range logs {
		t.Run(fmt.Sprintf( "Строка %v", i+1), func(t *testing.T) {
			if !FSM(str) && !(i == 3 || i == 2){
				t.Error("Не пройден FSM")
				return
			} else if FSM(str) && (i == 3 || i == 2){
				t.Error("Ошибочно пройден FSM")
				return
			}

			data := formatter.Format(str)

			// Обязательные поля
			if i != 2 && i != 3 && i != 5 {
				if _, ok := data["time"]; !ok {
					t.Error("Отсутствует свойство time")
				}
				if _, ok := data["event"]; !ok {
					t.Error("Отсутствует свойство event")
				}
				if v, ok := data["duration"]; !ok {
					t.Error("Отсутствует свойство duration")
				} else if v != "1003" {
					t.Errorf("Некорректное свойство duration, ожидалось 1003, имеем %v", v)
				}
			}
			if i == 4 {
				if min, ok := data["minutes"]; !ok || min != "38" {
					t.Error("Отсутствует или не верно определено свойство \"minutes\"")
				}
				if sec, ok := data["seconds"]; !ok || sec != "51" {
					t.Error("Отсутствует или не верно определено свойство \"seconds\"")
				}
			}
			if i == 5 {
				if _, ok := data["minutes"]; ok {
					t.Error("Присутствует свойство \"minutes\", должно отсутствовать")
				}
				if _, ok := data["seconds"]; ok {
					t.Error("Присутствует свойство \"seconds\", должно отсутствовать")
				}
			}

			if i == 0 {
				if len(data) != 22 {
					t.Errorf("Некорретное разбиение, должно быть 22 частей, имеем %v", len(data))
				}
				if data["Sql"] != "'SELECT Creation,Modified,Attributes,DataSize,BinaryData FROM Config WHERE FileName = $1 ORDER BY PartNo'" {
					t.Error("Некорректно распарсилось свойство \"Sql\"")
				}

				// проверяем ситуацию с вложенными кавычками, например
				// Prm="p_1: 'a7fe0aa5-722c-4a09-889b-ee422421addc'::mvarchar"
				if data["Prm"] != "\"p_1: 'a7fe0aa5-722c-4a09-889b-ee422421addc'::mvarchar\"" {
					t.Error("Некорректно распарсилось свойство \"Prm\"")
				}
			}
			if (i == 1 || i == 6) && len(data) != 20 {
				t.Errorf("Некорретное разбиение, должно быть 20 частей, имеем %v", len(data))
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