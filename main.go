package main

import (
	"bufio"
	"os"

	"context"
	"crypto/md5"
	"fmt"
	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
	"math"
	"math/rand"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	// _ "net/http/pprof"
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
	sLine       bool
	aggr, group string
	kp          *kingpin.Application
)

const (
	maxtableRow = 500 // Нет смысла выводить все строки, поставил 500. Если без оганичения тормаза начинаются
	modeDefault = 1 << iota
	modeSelect
	modeView
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())

	kp = kingpin.New("LogViewer", "Приложение для просмотра логов в табличном виде")
	kp.Flag("group", "Имена свойств для по которым нужно группировать").Short('g').StringVar(&group)
	kp.Flag("aggregate", "Имя свойства для агрегации (сумма, макс, ср)").Short('a').StringVar(&aggr)
	kp.Flag("savelines", "Если true значит уприложение будет сохранять исходные строки, что бы можно было посмотреть что вошло в ту или иную группировку. " +
		"Требует много оперативной памяти.").Short('s').Default("false").BoolVar(&sLine)

	runtime.SetMutexProfileFraction(5)
}

func main() {
	kp.Parse(os.Args[1:])

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeNamedPipe) == 0 {
		fmt.Println("Приложение работает только с pipe")
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

	//go http.ListenAndServe(":8888", nil)
	//go tool pprof  http://localhost:8888/debug/pprof/profile?seconds=10

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

	this.pages.AddPage("table", this.table, true, true)
	frame := tview.NewFrame(this.pages).SetBorders(0, 0, 0, 1, 0, 0)
	this.renderTableFooter(frame, modeDefault)

	textView := tview.NewTextView(). //.SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true).
		SetRegions(true).
		SetChangedFunc(func() {
			this.app.Draw()
		})
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
				this.renderTableFooter(frame, modeDefault)
			}
		}
		if key == tcell.KeyEnter {
			selectMode = true
			this.table.SetSelectable(selectMode, selectMode)
			this.renderTableFooter(frame, modeSelect)
		}
	}).SetSelectedFunc(func(row int, column int) {
		selectMode = false
		this.table.SetSelectable(selectMode, selectMode)
		clipboard.WriteAll(tview.TranslateANSI(this.table.GetCell(row, column).Text))
		this.renderTableFooter(frame, modeDefault)
	})
	//this.table.SetSelectionChangedFunc(func(row, column int) {
	//fmt.Println(1)
	//})
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
		if event.Key() == tcell.KeyTAB && !viewerMode { // почему-то tcell.KeyF2 не работает в linux
			row, _ := this.table.GetSelection()
			if !selectMode { // значит не вошли в режим выделения
				return event
			}
			column := this.table.GetColumnCount() - 1
			id := this.table.GetCell(row, column).Text

			viewerMode = true
			this.pages.AddPage("viewer", textView, true, true)
			textView.Clear()
			go func() {
				textView.ScrollToBeginning()
				if v, ok := this.line[id]; ok {
					txt := fmt.Sprintf(`["all"]%v[""]`, strings.Join(v.sourceLines, "\n")) //долго грузится при больших объемах
					textView.SetText(txt)

					//txt := append([]string{ `["all"]` }, append(v.sourceLines,  `[""]` )... )
					//for _, line := range txt {
					//	fmt.Fprintln(textView, line)  // append
					//	time.Sleep(time.Millisecond*10)
					//	textView.ScrollToBeginning()
					//}
				}
			}()

			this.renderTableFooter(frame, modeView)
		} else if event.Key() == tcell.KeyEscape {
			viewerMode = false
			this.pages.RemovePage("viewer")
			this.renderTableFooter(frame, modeSelect)
		}
		if event.Key() == tcell.KeyUp {
			row, _ := this.table.GetSelection()
			if row == 1 {
				return nil
			}
		}
		if event.Key() == tcell.KeyPgUp || event.Key() == tcell.KeyHome {
			//this.table.Select(1, 0)
			this.table.ScrollToBeginning()
			return nil
		}
		if event.Key() == tcell.KeyEnd {
			//this.table.Select(this.table.GetRowCount()-1 , 0)
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
				textView.Highlight()
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

	// Проверяем работу буфера, в линуксе его работа зависит от установленых приложений
	//if _, err := clipboard.ReadAll(); err != nil {
	//	frame.AddText(fmt.Sprintf("Произошла ошибка при работе с буфером обмена: %v", err), false, tview.AlignLeft, tcell.ColorRed)
	//}

	//this.pages.AddPage("footer", footer, true, true)
	if err := this.app.SetRoot(frame, true).EnableMouse(true).Run(); err != nil {
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
			count:       1,
		}
		if sLine {
			linedata.sourceLines = []string{line}
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
			if sLine {
				this.line[key].sourceLines = append(this.line[key].sourceLines, line)
			}

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

	// что б видно было по какой колонке отсортировано
	for i := 0; i < this.table.GetColumnCount(); i++ {
		if i == this.sortColumn && i <= startColCount-1 {
			this.table.GetCell(0, i).Text += " ▼"
		} else {
			this.table.GetCell(0, i).Text = strings.Replace(this.table.GetCell(0, i).Text, " ▼", "", -1)
		}
	}

}

func (this *tableView) renderTableFooter(footer *tview.Frame, mode int) {
	footer.Clear()
	if mode&modeDefault == modeDefault {
		footer.AddText("Exit - Esc", false, tview.AlignLeft, tcell.ColorGreen).
			AddText("Select mode - Enter", false, tview.AlignCenter, tcell.ColorGreen)

	}
	if mode&modeSelect == modeSelect {
		footer.AddText("Exit mode - Esc", false, tview.AlignLeft, tcell.ColorGreen).
			AddText("Copy in clipboard - Enter", false, tview.AlignCenter, tcell.ColorGreen).
			AddText("View lines - Tab", false, tview.AlignRight, tcell.ColorGreen)
	}
	if mode&modeView == modeView {
		footer.AddText("Exit view - Esc", false, tview.AlignLeft, tcell.ColorGreen).
			AddText("Copy in clipboard - Enter", false, tview.AlignCenter, tcell.ColorGreen)
	}
	// fmt.Printf("%-50v", "текст") - не работает с frame

	//frame.AddText(appendletter("Включить режим выделения строк", ".", 60) + "Enter", false, tview.AlignLeft, tcell.ColorBlue).
	//	AddText(appendletter("Скопировать значение ячейки в буфер (в режиме выделения)", ".", 60) + "Enter", false, tview.AlignLeft, tcell.ColorBlue).
	//	AddText(appendletter("Посмотреть исходные строки (в режиме выделения)", ".", 60) + "Tab", false, tview.AlignLeft, tcell.ColorBlue).
	//	AddText(appendletter("Выйти из режима выделения и из программы ", ".", 60) + "Esc", false, tview.AlignLeft, tcell.ColorBlue).
	//	SetBackgroundColor(tcell.ColorGreen)

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
		defer this.table.ScrollToBeginning()

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

func appendletter(str, letter string, count int) string {
	if len([]rune(str)) >= count {
		return str
	}

	delta := count - len([]rune(str))
	return str + strings.Repeat(letter, delta)
}
