package util

import (
	"encoding/json"
	"fmt"
	"github.com/jedib0t/go-pretty/table"
	"github.com/jedib0t/go-pretty/text"
	"io"
)

func JsonToString(jsonMap map[string]interface{}, raw bool) (string, error) {
	if !raw {
		prettyJSON, err := json.MarshalIndent(jsonMap, "", "\t")
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s", string(prettyJSON)), nil
	}
	rawJSON, err := json.Marshal(jsonMap)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s", string(rawJSON)), nil
}

func OutputTable(writer io.Writer, t *Table, csv bool) {
	tw := table.NewWriter()
	tw.AppendHeader(table.Row(t.Header.Row))
	for _, row := range t.Body {

		// Format floats
		for i, cd := range t.Header.ColumnDescriptors {
			if cd.ColumnType == ColumnType_Float && row[i] != "" {
				row[i] = fmt.Sprintf("%.5f", row[i])
			}
		}
		tw.AppendRow(table.Row(row))
	}
	tw.SetOutputMirror(writer)
	tw.SetAutoIndex(true)
	tw.SetStyle(table.StyleLight)
	columnConfigs := make([]table.ColumnConfig, 0)
	for i := range t.Header.Row {
		a := text.AlignLeft
		cds := t.Header.ColumnDescriptors
		if cds[i].ColumnType == ColumnType_Integer || cds[i].ColumnType == ColumnType_Float {
			a = text.AlignRight
		}
		columnConfigs = append(columnConfigs,
			table.ColumnConfig{Number: i + 1, Align: a})
	}
	tw.SetColumnConfigs(columnConfigs)

	// Finally, output the table.
	if csv {
		tw.RenderCSV()
	} else {
		tw.Render()
	}
}

//func (c *queryLogsCmd) writeChartDataSeries(dataSeries []sumologic.VisualDataSeries) {
//
//	plot.DefaultFont = "Helvetica"
//	p, err := plot.New()
//	if err != nil {
//		log.Fatal().Err(err).Msg("Error creating plot")
//	}
//
//	//p.Title.Text = "..."
//	p.X.Tick.Marker = plot.TimeTicks{Format: "2006-01-02\n15:04"}
//	//p.Y.Label.Text = "..."
//	p.Y.Tick.Marker = plot.DefaultTicks{}
//	p.Add(plotter.NewGrid())
//
//	for _, series := range dataSeries {
//		data := make(plotter.XYs, len(series.DataPoints))
//		for i, point := range series.DataPoints {
//			data[i].X = point.X / 1000
//			data[i].Y, err = strconv.ParseFloat(point.Y, 64)
//			if err != nil {
//				log.Fatal().Err(err).Str("float", point.Y).
//					Msg("Error converting DataPoints.Y string to float when plotting")
//			}
//		}
//		line, err := plotter.NewLine(data)
//		if err != nil {
//			log.Fatal().Err(err)
//		}
//		line.Color = color.RGBA{G: 255, A: 255}
//		p.Add(line)
//	}
//
//	err = p.Save(20 * vg.Centimeter, 10 * vg.Centimeter, "timeseries.png")
//	if err != nil {
//		log.Fatal().Err(err).Msg("Error saving chart image")
//	}
//
//	imgcat.CatFile("timeseries.png", os.Stderr)
//}
