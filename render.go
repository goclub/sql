package sq

import (
	"fmt"
	xerr "github.com/goclub/error"
	prettyTable "github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/jmoiron/sqlx"
	"strconv"
	"time"
)

var renderStyle prettyTable.Style

func init() {
	renderStyle = prettyTable.StyleColoredBlackOnBlueWhite
	renderStyle.Color.Header = text.Colors{text.BgHiBlack, text.FgHiWhite}
	renderStyle.Format.Header = text.FormatDefault
	renderStyle.Color.RowAlternate = renderStyle.Color.Row
}
func renderLastQueryCost(debugID uint64, lastQueryCost float64) (render string) {
	t := prettyTable.NewWriter()
	t.AppendHeader(prettyTable.Row{"LastQueryCost" + " (" + strconv.FormatUint(debugID, 10) + ")"})
	t.AppendRow([]interface{}{strconv.FormatFloat(lastQueryCost, 'f', -1, 64)})
	t.SetStyle(renderStyle)
	return "\n" + t.Render()
}
func renderRunTime(debugID uint64, duration time.Duration) (render string) {
	t := prettyTable.NewWriter()
	t.AppendHeader(prettyTable.Row{"RunTime" + " (" + strconv.FormatUint(debugID, 10) + ")"})
	t.AppendRow([]interface{}{duration.String()})
	t.SetStyle(renderStyle)
	return "\n" + t.Render()
}
func renderSQL(debugID uint64, query string, values []interface{}) (render string) {
	var printValues string
	for _, value := range values {
		printValues = printValues + fmt.Sprintf("%T(%v) ", value, value) + " "
	}
	t := prettyTable.NewWriter()
	t.AppendHeader(prettyTable.Row{"PrintSQL" + " (" + strconv.FormatUint(debugID, 10) + ")"}, prettyTable.RowConfig{AutoMerge: true})
	t.AppendRow(prettyTable.Row{query}, prettyTable.RowConfig{AutoMerge: true})
	t.AppendRow([]interface{}{printValues})
	t.SetStyle(renderStyle)
	return "\n" + t.Render()
}
func renderExplain(debugID uint64, row *sqlx.Row) (render string) {
	var err error
	defer func() {
		if err != nil {
			xerr.PrintStack(err)
		}
	}()
	t := prettyTable.NewWriter()
	t.AppendHeader(prettyTable.Row{"Explain" + " (" + strconv.FormatUint(debugID, 10) + ")"}, prettyTable.RowConfig{AutoMerge: true})
	t.AppendRow(prettyTable.Row{"id", "select_type", "table", "partitions",
		"type", "possible_keys", "key", "key_len", "ref",
		"rows", "filtered", "Extra"})
	var id, selectType, table, partitions, ttype, possibleLeys, key, keyLen, ref, rows, filtered, Extra *string
	err = row.Scan(&id, &selectType, &table, &partitions,
		&ttype, &possibleLeys, &key, &keyLen, &ref,
		&rows, &filtered, &Extra)
	_, err = CheckRowScanErr(err)
	if err != nil {
		return
	}
	sn := func(v *string) string {
		if v == nil {
			return "NULL"
		}
		return *v
	}
	t.AppendRow(prettyTable.Row{
		sn(id), sn(selectType), sn(table), sn(partitions),
		sn(ttype), sn(possibleLeys), sn(key), sn(keyLen), sn(ref),
		sn(rows), sn(filtered), sn(Extra),
	})
	t.SetStyle(renderStyle)
	return "\n" + t.Render()
}
func renderReview(debugID uint64, query string, reviews []string, refs string) (render string) {
	{
		t := prettyTable.NewWriter()
		t.SetStyle(renderStyle)
		t.AppendHeader(prettyTable.Row{"Review" + " (" + strconv.FormatUint(debugID, 10) + ")"}, prettyTable.RowConfig{AutoMerge: true})
		render = render + "\n" + t.Render()
	}
	{
		t := prettyTable.NewWriter()
		t.SetStyle(renderStyle)
		t.AppendHeader(prettyTable.Row{"Execute"}, prettyTable.RowConfig{AutoMerge: true})
		t.AppendRow(prettyTable.Row{query}, prettyTable.RowConfig{AutoMerge: true})
		render = render + "\n" + t.Render()
	}
	{
		t := prettyTable.NewWriter()
		t.SetStyle(renderStyle)
		t.AppendHeader(prettyTable.Row{"Reviews"}, prettyTable.RowConfig{AutoMerge: true})
		for _, review := range reviews {
			t.AppendRow(prettyTable.Row{review})
		}
		render = render + "\n" + t.Render()
	}
	return render
}
