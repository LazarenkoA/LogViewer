package main

import (
	"bufio"
	"context"
	"crypto/md5"
	"flag"
	"fmt"
	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"math"
	"math/rand"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type tline struct {
	count, summ, max, avg int

	//// служебное поле
	//avgSumm int

	// ключи
	keys        []string
	sourceLines []string
	id          string
}

type Iformatter interface {
	Format(string) map[string]string
}

type formatter1C struct{}

type tableView struct {
	sync.RWMutex

	app        *tview.Application
	table      *tview.Table
	pages      *tview.Pages
	line       map[string]*tline
	in         chan string
	ctx        context.Context
	formatter  Iformatter
	sortColumn int
}

var (
	showHelp    bool
	aggr, group string
	//max, summ, avg int
)

const (
	maxtableRow = 500 // Нет смысла выводить все строки, поставил 500. Если без оганичения тормаза начинаются
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	//flag.StringVar(&splitter, "s", defaultsplitter, "(splitter) Разделитель")
	flag.StringVar(&group, "g", "", "(group) Имена свойств для по которым нужно группировать")
	flag.StringVar(&aggr, "a", "", "(aggregate) Имя свойства для агрегации (сумма, макс, ср)")
	//flag.IntVar(&summ, "min", -1, "Порядковый номер числового поля для суммирования значения")
	//flag.IntVar(&max, "max", -1, "Порядковый номер числового поля для получения максимума")
	//flag.IntVar(&avg, "avg", -1, "Порядковый номер числового поля для получения среднего значения")
	flag.BoolVar(&showHelp, "help", false, "Помощь")
}

func main() {
	flag.Parse()
	if showHelp {
		flag.Usage()
		return
	}

	newView := new(tableView).Construct(context.Background())
	reader := bufio.NewReader(os.Stdin)
	go func() {
		for {
			if line, _, err := reader.ReadLine(); err == nil {
				newView.in <- string(line)
			} else {
				close(newView.in)
				go newView.showmodal("Готово")
				break
			}
		}
	}()

	//source := []string{
	//	`00:00.953001-0,CLSTR,0,process=ragent,OSThread=12746,host:pid=ca-sys-3:5822,current_rt=46,average_rt=51`,
	//	`00:00.976002-0,CLSTR,0,process=ragent,OSThread=13138,host:pid=ca-sys-3:5824,current_rt=48,average_rt=51`,
	//	`00:00.982005-0,CLSTR,0,process=ragent,OSThread=12747,host:pid=CA-T3-APP-1:12754,current_rt=54,average_rt=43`,
	//	`00:01.149000-0,CONN,0,process=ragent,OSThread=12749,Txt='Ping direction statistics: address=127.0.1.1:1560,pingTimeout=15000,pingPeriod=3000,period=10065,packetsSent=3,avgResponseTime=0,maxResponseTime=0,packetsTimedOut=0,packetsLost=0,packetsLostAndFound=0'`,
	//	`00:01.150000-0,CONN,0,process=ragent,OSThread=12749,Txt='Ping direction statistics: address=172.18.1.27:31562,pingTimeout=15000,pingPeriod=3000,period=10065,packetsSent=3,avgResponseTime=0,maxResponseTime=0,packetsTimedOut=0,packetsLost=0,packetsLostAndFound=0'`,
	//	`00:01.150001-0,CONN,0,process=ragent,OSThread=12749,Txt='Ping direction statistics: address=172.18.1.27:31563,pingTimeout=15000,pingPeriod=3000,period=10065,packetsSent=3,avgResponseTime=0,maxResponseTime=0,packetsTimedOut=0,packetsLost=0,packetsLostAndFound=0'`,
	//	`00:02.220000-0,CONN,0,process=ragent,OSThread=12855,ClientID=11211,Protected=0,Txt='Accepted, client=(2)127.0.0.1:48946, server=(2)127.0.0.1:1540'`,
	//	`00:02.221000-0,CONN,0,process=ragent,OSThread=12857,ClientID=11212,Protected=0,Txt='Accepted, client=(2)127.0.0.1:48948, server=(2)127.0.0.1:1540'`,
	//	`00:02.221009-0,CONN,0,process=ragent,OSThread=12859,ClientID=11213,Protected=0,Txt='Accepted, client=(2)127.0.0.1:48950, server=(2)127.0.0.1:1540'`,
	//}
	//
	//
	//go func() {
	//	t := time.NewTicker(time.Millisecond)
	//	defer close(newView.in)
	//	defer t.Stop()
	//
	//	for {
	//		newView.in <- source[rand.Intn(len(source))]
	//
	//		<- t.C
	//	}
	//}()

	// cat D:/log/T3/For1C/rphost*/* | grep CALL -w | grep context -iw | main -a=Memory -g=event,Context
	newView.start()
}

func (this *tableView) Construct(ctx context.Context) *tableView {
	this.in = make(chan string, 5)
	this.ctx = ctx
	this.app = tview.NewApplication()
	this.table = tview.NewTable().SetBorders(false).SetFixed(0, 0)
	this.pages = tview.NewPages()
	this.line = make(map[string]*tline, 0)
	// this.sortColumn = -1 // если оставить -1 то строки не будут по дефолту сортироваться

	return this
}

func (this *tableView) start() {
	this.tableHeader()
	go this.tableFill()

	textView := tview.NewTextView().SetDynamicColors(true).SetScrollable(true).SetWordWrap(true).SetRegions(true)
	textView.SetBorder(true)

	selectMode, viewerMode := false, false
	// события таблицы
	this.table.Select(1, 0).SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			if !selectMode {
				this.app.Stop()
			} else {
				selectMode = false
				this.table.SetSelectable(selectMode, selectMode)
			}
		}
		if key == tcell.KeyEnter {
			selectMode = true
			this.table.SetSelectable(selectMode, selectMode)
		}
	}).SetSelectedFunc(func(row int, column int) {
		selectMode = false
		this.table.SetSelectable(selectMode, selectMode)
		clipboard.WriteAll(tview.TranslateANSI(this.table.GetCell(row, column).Text))
	})
	this.table.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (action2 tview.MouseAction, mouse *tcell.EventMouse) {
		// пока не нашел другого способа понять по какой ячейке кликнули
		if action == tview.MouseLeftClick && this.table.GetColumnCount() > 1 {
			mouseX, mouseY := event.Position()
			var column int = -1
			for i := 0; i < this.table.GetColumnCount(); i++ {
				x, _, width := this.table.GetCell(0, i).GetLastPosition() // колонки шапки
				if mouseX > x && mouseX < x+width && mouseY == 0 {        // Y странный, это номер строки по которой кликнули
					//cell = this.table.GetCell(0, i)
					column = i
				}
			}
			if column != -1 {
				this.sortColumn = column
				this.forceRenderTable()
			}
		}

		return action, event
	})
	this.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTAB  && !viewerMode { // почему-то tcell.KeyF2 не работает в linux
			row, _ := this.table.GetSelection()
			if !selectMode { // значит не вошли в режим выделения
				return event
			}
			column := this.table.GetColumnCount()-1
			id := this.table.GetCell(row, column).Text

			if v, ok := this.line[id]; ok {
				txt := fmt.Sprintf(`["all"]%v[""]`, strings.Join(v.sourceLines, "\n"))
				//fmt.Fprintf(textView, "%s ", txt) // append
				textView.SetText(txt)
				textView.ScrollToBeginning()
			}

			viewerMode = true
			go func() {
				this.pages.AddPage("viewer", textView, true, true)
				this.app.Draw()
			}()

		} else if event.Key() == tcell.KeyEscape {
			viewerMode = false
			this.pages.RemovePage("viewer")
		}
		if event.Key() == tcell.KeyUp {
			row, _ := this.table.GetSelection()
			if row == 1 {
				return nil
			}
		}
		if event.Key() == tcell.KeyPgUp || event.Key() == tcell.KeyHome {
			this.table.Select(1, 0)
			this.table.ScrollToBeginning()
			return nil
		}
		if event.Key() == tcell.KeyEnd {
			this.table.Select(this.table.GetRowCount()-1 , 0)
			this.table.ScrollToEnd()
			return nil
		}
		return event
	})


	// события TextView
	selectMode = false
	textView.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			go func() {
				textView.Highlight("all").ScrollToHighlight()
				this.app.Draw()
				clipboard.WriteAll(textView.GetText(true))
				time.Sleep(time.Millisecond * 100)
				textView.Highlight()
				this.app.Draw()
			}()
		}
	})

	this.app.SetFocus(this.table)
	//flex := tview.NewFlex().
	//	AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
	//		AddItem(this.table, 0, 1, false).
	//		AddItem(tview.NewBox().SetBorder(true).SetTitle("Bottom (5 rows)"). , 5, 1, false), 0, 2, false)

	frame := tview.NewFrame(this.table).SetBorders(0, 2, 0, 2, 1, 2)

	// Проверяем работу буфера, в линуксе его работа зависит от установленых приложений
	//if _, err := clipboard.ReadAll(); err != nil {
	//	frame.AddText(fmt.Sprintf("Произошла ошибка при работе с буфером обмена: %v", err), false, tview.AlignLeft, tcell.ColorRed)
	//}

	this.pages.AddPage("frame", frame, true, true)
	if err := this.app.SetRoot(this.pages, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

func (this *tableView) tableFill() {
	t := time.NewTicker(time.Millisecond * 500) // каждую 1/2 секунду обновляем таблицу
	go func() {
		for {
			this.renderTable()
			<-t.C
		}
	}()

	formatter := new(formatter1C)
	for line := range this.in {
		linedata := &tline{
			keys:        []string{},
			sourceLines: []string{line},
			count:       1,
		}

		fline := formatter.Format(line)

		groupField := strings.Split(group, ",")
		if len(groupField) > 0 {
			for _, field := range groupField {
				linedata.keys = append(linedata.keys, fline[field])
			}
		} else {
			linedata.keys = append(linedata.keys, line)
		}

		var intVal int
		if aggr != "" {
			intVal, _ = strconv.Atoi(fline[aggr])
		}

		key := getHash(strings.Join(linedata.keys, "-"))
		linedata.id = key
		if this.lineExist(key) {
			this.line[key].count++
			this.line[key].sourceLines = append(this.line[key].sourceLines, line)
			this.line[key].max = int(math.Max(float64(intVal), float64(this.line[key].max)))
			this.line[key].summ += intVal
			this.line[key].avg = this.line[key].summ / this.line[key].count
		} else {
			linedata.max = int(math.Max(float64(intVal), float64(linedata.max)))
			linedata.summ += intVal
			this.Addline(key, linedata)
		}
	}

	// останавливаем таймер и рендарим что б вывести хвосты
	t.Stop()
	this.renderTable()
}

func (this *tableView) tableHeader() {
	this.table.SetCell(0, 0, tview.NewTableCell("Количество").
		SetTextColor(tcell.ColorWhite).
		SetBackgroundColor(tcell.ColorGreen).
		SetAlign(tview.AlignLeft).
		SetSelectable(false))

	if aggr != "" {
		this.table.SetCell(0, 1, tview.NewTableCell(fmt.Sprintf("Суммма (%v)", aggr)).
			SetTextColor(tcell.ColorWhite).
			SetBackgroundColor(tcell.ColorGreen).
			SetAlign(tview.AlignLeft).
			SetSelectable(false))
		this.table.SetCell(0, 2, tview.NewTableCell(fmt.Sprintf("Максимум (%v)", aggr)).
			SetTextColor(tcell.ColorWhite).
			SetBackgroundColor(tcell.ColorGreen).
			SetAlign(tview.AlignLeft).
			SetSelectable(false))
		this.table.SetCell(0, 3, tview.NewTableCell(fmt.Sprintf("Среднее (%v)", aggr)).
			SetTextColor(tcell.ColorWhite).
			SetBackgroundColor(tcell.ColorGreen).
			SetAlign(tview.AlignLeft).
			SetSelectable(false))
	}

	startColCount := this.table.GetColumnCount()
	for i, v := range strings.Split(group, ",") {
		this.table.SetCell(0, startColCount+i, tview.NewTableCell(v).
			SetTextColor(tcell.ColorWhite).
			SetBackgroundColor(tcell.ColorGreen).
			SetAlign(tview.AlignLeft).
			SetSelectable(false))
	}
}

func (this *tableView) Addline(key string, value *tline) {
	this.Lock()
	defer this.Unlock()

	this.line[key] = value
}

func (this *tableView) lineExist(key string) bool {
	this.RLock()
	defer this.RUnlock()

	_, ok := this.line[key]
	return ok
}

func (this *tableView) forceRenderTable() {
	this.table.Clear()
	this.tableHeader()
	this.renderTable()
}

func (this *tableView) renderTable() {
	this.RLock()
	defer this.RUnlock()
	defer this.table.ScrollToBeginning()

	// перекладываем из мапы в массив, что б его потом сортировать
	dataArray := []*tline{}
	for _, v := range this.line {
		dataArray = append(dataArray, v)
	}
	if this.sortColumn >= 0 {
		sort.Slice(dataArray, func(i, j int) bool {
			switch this.sortColumn {
			case 1:
				return dataArray[i].summ > dataArray[j].summ
			case 2:
				return dataArray[i].max > dataArray[j].max
			case 3:
				return dataArray[i].avg > dataArray[j].avg
			default:
				return dataArray[i].count > dataArray[j].count
			}
		})
	}

	go this.app.QueueUpdateDraw(func() {
	continueLine:
		for _, v := range dataArray {
			row := this.table.GetRowCount()

			if row >= maxtableRow {
				break
			}

			// агрегируемые поля
			// обновление данных в сущ. строках
			for i := 1; i < this.table.GetRowCount(); i++ {
				id := this.table.GetCell(i, this.table.GetColumnCount()-1).Text // в последней колонке идентификатор строки
				if id == v.id {
					this.table.GetCell(i, 0).SetText(strconv.Itoa(v.count))
					if aggr != "" {
						this.table.GetCell(i, 1).SetText(strconv.Itoa(v.summ))
						this.table.GetCell(i, 2).SetText(strconv.Itoa(v.max))
						this.table.GetCell(i, 3).SetText(strconv.Itoa(v.avg))
					}
					continue continueLine
				}
			}

			// все что ниже это добавление новой строки
			cCount := 1
			this.table.SetCell(row, 0, tview.NewTableCell(strconv.Itoa(v.count)).
				SetTextColor(tcell.ColorWhite).
				SetAlign(tview.AlignLeft))
			if aggr != "" {
				this.table.SetCell(row, 1, tview.NewTableCell(strconv.Itoa(v.summ)).
					SetTextColor(tcell.ColorWhite).
					SetAlign(tview.AlignLeft))
				this.table.SetCell(row, 2, tview.NewTableCell(strconv.Itoa(v.max)).
					SetTextColor(tcell.ColorWhite).
					SetAlign(tview.AlignLeft))
				this.table.SetCell(row, 3, tview.NewTableCell(strconv.Itoa(v.avg)).
					SetTextColor(tcell.ColorWhite).
					SetAlign(tview.AlignLeft))
				cCount = 4
			}

			// группируемые поля
			for i, v := range v.keys {
				this.table.SetCell(row, cCount+i, tview.NewTableCell(v).
					SetTextColor(tcell.ColorWhite).
					SetAlign(tview.AlignLeft))
			}

			// ключ строки (по нему далее будет обновляться значения агрегируемых полей)
			this.table.SetCell(row, len(v.keys)+cCount, tview.NewTableCell(v.id).SetSelectable(false).
				SetTextColor(this.table.GetBackgroundColor()).
				SetMaxWidth(1))
		}
	})
}

func (this *tableView) showmodal(str string) {
	modal := tview.NewModal().
		SetText(str).
		AddButtons([]string{"Ок"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Ок" {
				this.pages.RemovePage("modal")
				this.app.SetFocus(this.table)
			}
		})
	this.app.SetFocus(modal)
	this.pages.AddPage("modal", modal, true, true)
	this.app.ForceDraw()
	//this.pages.Draw()
}

func getHash(inStr string) string {
	Sum := md5.Sum([]byte(inStr))
	return fmt.Sprintf("%x", Sum)
}

func (f *formatter1C) Format(str string) map[string]string {
	result := make(map[string]string, 0)

	// проверяем на соответствие шаблону, важно при обработке многострочных логов
	re := regexp.MustCompile(`(?mi)\d\d:\d\d\.\d+[-]\d+`)
	if ok := re.MatchString(str); !ok {
		return result
	}

	parts := strings.Split(str, ",")
	if len(parts) == 0 {
		return result
	}
	for _, v := range parts {
		keyValue := strings.Split(strings.Trim(v, " "), "=")
		if len(keyValue) == 2 {
			result[keyValue[0]] = keyValue[1]
		}
	}

	// теперь системные свойства, время, событие, длительность (06:11.062003-0,CLSTR,0,pro....)
	// время
	result["time"] = parts[0][:strings.Index(parts[0], "-")]

	// длительность
	result["duration"] = parts[0][strings.Index(parts[0], "-")+1:]

	// событие
	result["event"] = parts[1]

	return result
}
