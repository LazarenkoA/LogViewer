package main

import (
	"bufio"
	"os"

	"context"
	"crypto/md5"
	"encoding/csv"
	"fmt"
	"math"
	"math/rand"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
	// _ "net/http/pprof"
)

var messages = map[string]map[string]string{
	"ru": {
		"appName":           "Приложение для просмотра логов в табличном виде. https://github.com/LazarenkoA/LogViewer",
		"group":             "Имена свойств для по которым нужно группировать",
		"aggregate":         "Имя свойства для агрегации (сумма, макс, ср)",
		"savelines":         "Если true значит приложение будет сохранять исходные строки, что бы можно было посмотреть что вошло в ту или иную группировку. Требует много оперативной памяти.",
		"pipe":              "Приложение работает только с pipe",
		"done":              "Готово",
		"count":             "Количество",
		"summ":              "Сумма (%v)",
		"max":               "Максимум (%v)",
		"avg":               "Среднее (%v)",
		"exit":              "Выход - Esc",
		"selectMode":        "Режим выбора - Enter",
		"exportToCSV":       "Экспорт в CSV - F5",
		"copyToClipboard":   "Копировать в буфер - Enter",
		"viewRowsDetails":   "Просмотр строк - Tab",
		"createFileError":   "Ошибка при создании файла: %v",
		"writeFileError":    "Ошибка при записи заголовков: %v",
		"writeFileSucccess": "Файл успешно сохранен в %s",
		"inputFileName":     "Имя файла: ",
		"save":              "Сохранить",
		"cancel":            "Отменить",
	},
	"en": {
		"appName":           "Application for view onec tech logs in table view. https://github.com/LazarenkoA/LogViewer",
		"group":             "Names of properties for grouping",
		"aggregate":         "Name of property for aggregation (sum, max, avg)",
		"savelines":         "If true, the application will save the original lines, so you can see what went into the group. Requires a lot of memory.",
		"pipe":              "The application only works with pipe",
		"done":              "Done",
		"count":             "Count",
		"summ":              "Sum (%v)",
		"max":               "Maximum (%v)",
		"avg":               "Average (%v)",
		"exit":              "Exit - Esc",
		"selectMode":        "Select mode - Enter",
		"exportToCSV":       "Export to CSV - F5",
		"copyToClipboard":   "Copy to clipboard - Enter",
		"viewRowsDetails":   "View rows - Tab",
		"createFileError":   "Error creating file: %v",
		"writeFileError":    "Error writing headers: %v",
		"writeFileSucccess": "File successfully saved in %s",
		"inputFileName":     "File name: ",
		"save":              "Save",
		"cancel":            "Cancel",
	},
}

type tline struct {
	count, summ, max, avg int

	//// служебное поле
	//avgSumm int

	// ключи
	keys        []string
	sourceLines []string
	id          string
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

	lang := detectLanguage()

	kp = kingpin.New("LogViewer", messages[lang]["appName"])
	kp.Flag("group", messages[lang]["group"]).Short('g').StringVar(&group)
	kp.Flag("aggregate", messages[lang]["aggregate"]).Short('a').StringVar(&aggr)
	kp.Flag("savelines", messages[lang]["savelines"]).Short('s').Default("false").BoolVar(&sLine)

	runtime.SetMutexProfileFraction(5)
}

func main() {
	lang := detectLanguage()

	kp.Parse(os.Args[1:])

	stat, _ := os.Stdin.Stat()

	if (stat.Mode() & os.ModeNamedPipe) == 0 {
		fmt.Println(messages[lang]["pipe"])
		return
	}

	newView := new(tableView).Construct(context.Background())
	newView.lang = lang
	reader := bufio.NewReader(os.Stdin)
	go func() {
		for {
			if line, _, err := reader.ReadLine(); err == nil {
				newView.in <- string(line)
			} else {
				close(newView.in)
				go newView.showmodal(messages[newView.lang]["done"])
				break
			}
		}
	}()

	//go http.ListenAndServe(":8888", nil)
	//go tool pprof  http://localhost:8888/debug/pprof/profile?seconds=10

	newView.start()
}

func (tv *tableView) Construct(ctx context.Context) *tableView {
	tv.in = make(chan string, 5)
	tv.ctx = ctx
	tv.app = tview.NewApplication()
	tv.table = tview.NewTable().SetBorders(false).SetFixed(0, 0)
	tv.pages = tview.NewPages()
	tv.line = make(map[string]*tline, 0)
	// this.sortColumn = -1 // если оставить -1 то строки не будут по дефолту сортироваться

	return tv
}

func (tv *tableView) start() {
	tv.tableHeader()
	go tv.tableFill()

	tv.pages.AddPage("table", tv.table, true, true)
	frame := tview.NewFrame(tv.pages).SetBorders(0, 0, 0, 1, 0, 0)
	tv.renderTableFooter(frame, modeDefault)

	textView := tview.NewTextView(). //.SetDynamicColors(true).
						SetScrollable(true).
						SetWordWrap(true).
						SetRegions(true).
						SetChangedFunc(func() {
			tv.app.Draw()
		})
	textView.SetBorder(true)

	selectMode, viewerMode := false, false
	// события таблицы
	tv.table.Select(1, 0).SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			if !selectMode {
				tv.app.Stop()
			} else {
				selectMode = false
				tv.table.SetSelectable(selectMode, selectMode)
				tv.renderTableFooter(frame, modeDefault)
			}
		}
		if key == tcell.KeyEnter {
			selectMode = true
			tv.table.SetSelectable(selectMode, selectMode)
			tv.renderTableFooter(frame, modeSelect)
		}
	}).SetSelectedFunc(func(row int, column int) {
		selectMode = false
		tv.table.SetSelectable(selectMode, selectMode)
		clipboard.WriteAll(tview.TranslateANSI(tv.table.GetCell(row, column).Text))
		tv.renderTableFooter(frame, modeDefault)
	})
	//this.table.GetCell(0, 1).SetClickedFunc(func()bool {
	//	fmt.Println(1)
	//	return true
	//})
	tv.table.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (action2 tview.MouseAction, mouse *tcell.EventMouse) {
		// пока не нашел другого способа понять по какой ячейке кликнули
		if action == tview.MouseLeftClick && tv.table.GetColumnCount() > 1 {
			mouseX, mouseY := event.Position()
			var column int = -1
			for i := 0; i < tv.table.GetColumnCount(); i++ {
				x, _, width := tv.table.GetCell(0, i).GetLastPosition() // колонки шапки
				if mouseX > x && mouseX < x+width && mouseY == 0 {      // Y странный, это номер строки по которой кликнули
					//cell = this.table.GetCell(0, i)
					column = i
				}
			}
			if column != -1 {
				tv.sortColumn = column
				tv.forceRenderTable()
			}
		}

		return action, event
	})
	tv.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTAB && !viewerMode { // почему-то tcell.KeyF2 не работает в linux
			row, _ := tv.table.GetSelection()
			if !selectMode { // значит не вошли в режим выделения
				return event
			}
			column := tv.table.GetColumnCount() - 1
			id := tv.table.GetCell(row, column).Text

			viewerMode = true
			tv.pages.AddPage("viewer", textView, true, true)
			textView.Clear()
			go func() {
				textView.ScrollToBeginning()
				if v, ok := tv.line[id]; ok {
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

			tv.renderTableFooter(frame, modeView)
		} else if event.Key() == tcell.KeyEscape {
			viewerMode = false
			tv.pages.RemovePage("viewer")
			tv.renderTableFooter(frame, modeSelect)
		}
		if event.Key() == tcell.KeyUp {
			row, _ := tv.table.GetSelection()
			if row == 1 {
				return nil
			}
		}
		if event.Key() == tcell.KeyPgUp || event.Key() == tcell.KeyHome {
			//this.table.Select(1, 0)
			tv.table.ScrollToBeginning()
			return nil
		}
		if event.Key() == tcell.KeyEnd {
			//this.table.Select(this.table.GetRowCount()-1 , 0)
			tv.table.ScrollToEnd()
			return nil
		}
		if event.Key() == tcell.KeyF5 {
			go tv.SaveToCSV()
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
				tv.app.Draw()
				clipboard.WriteAll(textView.GetText(true))
				time.Sleep(time.Millisecond * 100)
				textView.Highlight()
				tv.app.Draw()
			}()
		}
	})

	tv.app.SetFocus(tv.table)
	//flex := tview.NewFlex().
	//	AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
	//		AddItem(this.table, 0, 1, false).
	//		AddItem(tview.NewBox().SetBorder(true).SetTitle("Bottom (5 rows)"). , 5, 1, false), 0, 2, false)

	// Проверяем работу буфера, в линуксе его работа зависит от установленых приложений
	//if _, err := clipboard.ReadAll(); err != nil {
	//	frame.AddText(fmt.Sprintf("Произошла ошибка при работе с буфером обмена: %v", err), false, tview.AlignLeft, tcell.ColorRed)
	//}

	//this.pages.AddPage("footer", footer, true, true)
	if err := tv.app.SetRoot(frame, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

func (tv *tableView) tableFill() {
	t := time.NewTicker(time.Millisecond * 500) // каждую 1/2 секунду обновляем таблицу
	go func() {
		for {
			tv.renderTable()
			<-t.C
		}
	}()

	formatter := new(formatter1C)
	linedata := &tline{
		keys:  []string{},
		count: 1,
	}
	var sourceLine string

	for line := range tv.in {
		fline, err := formatter.Format(line)

		if err != nil {
			sourceLine += line + "\n"
			continue
		}
		sourceLine += line

		if sLine {
			linedata.sourceLines = []string{sourceLine}
		}

		sourceLine = ""

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
		if tv.lineExist(key) {
			tv.line[key].count++
			if sLine {
				tv.line[key].sourceLines = append(tv.line[key].sourceLines, linedata.sourceLines[0])
			}

			tv.line[key].max = int(math.Max(float64(intVal), float64(tv.line[key].max)))
			tv.line[key].summ += intVal
			tv.line[key].avg = tv.line[key].summ / tv.line[key].count
		} else {
			linedata.max = int(math.Max(float64(intVal), float64(linedata.max)))
			linedata.summ += intVal
			tv.Addline(key, linedata)
		}

		linedata = &tline{
			keys:  []string{},
			count: 1,
		}
	}

	// останавливаем таймер и рендарим что б вывести хвосты
	t.Stop()
	tv.renderTable()
}

func (tv *tableView) tableHeader() {
	tv.table.SetCell(0, 0, tview.NewTableCell(messages[tv.lang]["count"]).
		SetTextColor(tcell.ColorWhite).
		SetBackgroundColor(tcell.ColorGreen).
		SetAlign(tview.AlignLeft).
		SetSelectable(false))

	if aggr != "" {
		tv.table.SetCell(0, 1, tview.NewTableCell(fmt.Sprintf(messages[tv.lang]["summ"], aggr)).
			SetTextColor(tcell.ColorWhite).
			SetBackgroundColor(tcell.ColorGreen).
			SetAlign(tview.AlignLeft).
			SetSelectable(false))
		tv.table.SetCell(0, 2, tview.NewTableCell(fmt.Sprintf(messages[tv.lang]["max"], aggr)).
			SetTextColor(tcell.ColorWhite).
			SetBackgroundColor(tcell.ColorGreen).
			SetAlign(tview.AlignLeft).
			SetSelectable(false))
		tv.table.SetCell(0, 3, tview.NewTableCell(fmt.Sprintf(messages[tv.lang]["avg"], aggr)).
			SetTextColor(tcell.ColorWhite).
			SetBackgroundColor(tcell.ColorGreen).
			SetAlign(tview.AlignLeft).
			SetSelectable(false))
	}

	startColCount := tv.table.GetColumnCount()
	for i, v := range strings.Split(group, ",") {
		tv.table.SetCell(0, startColCount+i, tview.NewTableCell(v).
			SetTextColor(tcell.ColorWhite).
			SetBackgroundColor(tcell.ColorGreen).
			SetAlign(tview.AlignLeft).
			SetSelectable(false))
	}

	// что б видно было по какой колонке отсортировано
	for i := 0; i < tv.table.GetColumnCount(); i++ {
		if i == tv.sortColumn && i <= startColCount-1 {
			tv.table.GetCell(0, i).Text += " ▼"
		} else {
			tv.table.GetCell(0, i).Text = strings.Replace(tv.table.GetCell(0, i).Text, " ▼", "", -1)
		}
	}

}

func (tv *tableView) renderTableFooter(footer *tview.Frame, mode int) {
	footer.Clear()
	if mode&modeDefault == modeDefault {
		footer.AddText(messages[tv.lang]["exit"], false, tview.AlignLeft, tcell.ColorGreen).
			AddText(messages[tv.lang]["selectMode"], false, tview.AlignCenter, tcell.ColorGreen).
			AddText(messages[tv.lang]["exportToCSV"], false, tview.AlignRight, tcell.ColorGreen)
	}
	if mode&modeSelect == modeSelect {
		footer.AddText(messages[tv.lang]["exit"], false, tview.AlignLeft, tcell.ColorGreen).
			AddText(messages[tv.lang]["copyToClipboard"], false, tview.AlignCenter, tcell.ColorGreen).
			AddText(messages[tv.lang]["viewRowsDetails"], false, tview.AlignRight, tcell.ColorGreen)
	}
	if mode&modeView == modeView {
		footer.AddText(messages[tv.lang]["exit"], false, tview.AlignLeft, tcell.ColorGreen).
			AddText(messages[tv.lang]["copyToClipboard"], false, tview.AlignCenter, tcell.ColorGreen)
	}

}

func (tv *tableView) Addline(key string, value *tline) {
	tv.Lock()
	defer tv.Unlock()

	tv.line[key] = value
}

func (tv *tableView) lineExist(key string) bool {
	tv.RLock()
	defer tv.RUnlock()

	_, ok := tv.line[key]
	return ok
}

func (tv *tableView) forceRenderTable() {
	tv.table.Clear()
	tv.tableHeader()
	tv.renderTable()
}

func (tv *tableView) renderTable() {
	tv.RLock()
	defer tv.RUnlock()

	// перекладываем из мапы в массив, что б его потом сортировать
	dataArray := []*tline{}
	for _, v := range tv.line {
		dataArray = append(dataArray, v)
	}
	if tv.sortColumn >= 0 {
		sort.Slice(dataArray, func(i, j int) bool {
			switch tv.sortColumn {
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

	go tv.app.QueueUpdateDraw(func() {
		defer tv.table.ScrollToBeginning()

	continueLine:
		for _, v := range dataArray {
			row := tv.table.GetRowCount()

			if row >= maxtableRow {
				break
			}

			// агрегируемые поля
			// обновление данных в сущ. строках
			for i := 1; i < tv.table.GetRowCount(); i++ {
				id := tv.table.GetCell(i, tv.table.GetColumnCount()-1).Text // в последней колонке идентификатор строки
				if id == v.id {
					tv.table.GetCell(i, 0).SetText(strconv.Itoa(v.count))
					if aggr != "" {
						tv.table.GetCell(i, 1).SetText(strconv.Itoa(v.summ))
						tv.table.GetCell(i, 2).SetText(strconv.Itoa(v.max))
						tv.table.GetCell(i, 3).SetText(strconv.Itoa(v.avg))
					}
					continue continueLine
				}
			}

			// все что ниже это добавление новой строки
			cCount := 1
			tv.table.SetCell(row, 0, tview.NewTableCell(strconv.Itoa(v.count)).
				SetTextColor(tcell.ColorWhite).
				SetAlign(tview.AlignLeft))
			if aggr != "" {
				tv.table.SetCell(row, 1, tview.NewTableCell(strconv.Itoa(v.summ)).
					SetTextColor(tcell.ColorWhite).
					SetAlign(tview.AlignLeft))
				tv.table.SetCell(row, 2, tview.NewTableCell(strconv.Itoa(v.max)).
					SetTextColor(tcell.ColorWhite).
					SetAlign(tview.AlignLeft))
				tv.table.SetCell(row, 3, tview.NewTableCell(strconv.Itoa(v.avg)).
					SetTextColor(tcell.ColorWhite).
					SetAlign(tview.AlignLeft))
				cCount = 4
			}

			// группируемые поля
			for i, v := range v.keys {
				tv.table.SetCell(row, cCount+i, tview.NewTableCell(v).
					SetTextColor(tcell.ColorWhite).
					SetAlign(tview.AlignLeft))
			}

			// ключ строки (по нему далее будет обновляться значения агрегируемых полей)
			tv.table.SetCell(row, len(v.keys)+cCount, tview.NewTableCell(v.id).SetSelectable(false).
				SetTextColor(tv.table.GetBackgroundColor()).
				SetMaxWidth(1))

		}
	})
}

func (tv *tableView) showmodal(str string) {
	modal := tview.NewModal().
		SetText(str).
		AddButtons([]string{"Ок"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Ок" {
				tv.pages.RemovePage("modal")
				tv.app.SetFocus(tv.table)
			}
		})
	tv.app.SetFocus(modal)
	tv.pages.AddPage("modal", modal, true, true)
	tv.app.ForceDraw()
	//this.pages.Draw()
}

func getHash(inStr string) string {
	Sum := md5.Sum([]byte(inStr))
	return fmt.Sprintf("%x", Sum)
}

func (tv *tableView) exportToCSV() {
	// Запрашиваем имя файла у пользователя

	if tv.csvFileName == "" {
		return // Пользователь отменил операцию
	}

	// Добавляем расширение .csv, если его нет
	if !strings.HasSuffix(tv.csvFileName, ".csv") {
		tv.csvFileName += ".csv"
	}

	// Выполняем экспорт в отдельной горутине
	go func() {
		file, err := os.Create(tv.csvFileName)
		if err != nil {
			tv.showmodal(fmt.Sprintf(messages[tv.lang]["createFileError"], err))
			return
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		defer writer.Flush()

		// Записываем заголовки
		headers := make([]string, 0, tv.table.GetColumnCount())
		for col := 0; col < tv.table.GetColumnCount()-1; col++ {
			headers = append(headers, tv.table.GetCell(0, col).Text)
		}
		if err := writer.Write(headers); err != nil {
			tv.showmodal(fmt.Sprintf(messages[tv.lang]["writeFileError"], err))
			return
		}

		// Записываем данные
		for row := 1; row < tv.table.GetRowCount(); row++ {
			rowData := make([]string, 0, tv.table.GetColumnCount())
			for col := 0; col < tv.table.GetColumnCount()-1; col++ {
				rowData = append(rowData, tv.table.GetCell(row, col).Text)
			}
			if err := writer.Write(rowData); err != nil {
				tv.showmodal(fmt.Sprintf(messages[tv.lang]["writeFileError"], err))
				return
			}
		}

		tv.showmodal(fmt.Sprintf(messages[tv.lang]["writeFileSucccess"], tv.csvFileName))
	}()
}

func (tv *tableView) SaveToCSV() {

	inputField := tview.NewInputField().SetLabel(messages[tv.lang]["inputFileName"])

	tv.csvFileName = ""

	form := tview.NewForm().
		AddFormItem(inputField).
		AddButton(messages[tv.lang]["save"], func() {
			tv.csvFileName = inputField.GetText()
			tv.pages.RemovePage("saveToCSV")
			tv.app.SetFocus(tv.table)
			tv.exportToCSV()
		}).
		AddButton(messages[tv.lang]["cancel"], func() {
			tv.pages.RemovePage("saveToCSV")
			tv.app.SetFocus(tv.table)
		})

	form.SetBorder(true).SetTitle(messages[tv.lang]["exportToCSV"]).SetTitleAlign(tview.AlignCenter)

	tv.pages.AddPage("saveToCSV", form, true, true)
	tv.app.ForceDraw()
}

func detectLanguage() string {
	locale := os.Getenv("LANG")

	locale = strings.ToLower(locale)

	if strings.HasPrefix(locale, "ru") {
		return "ru"
	}

	return "en"
}
