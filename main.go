package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/ful09003/tinderbox/pkg/scrape"
	"github.com/ful09003/tinderbox/pkg/types"
	dto "github.com/prometheus/client_model/go"
)

func main() {
	lhsEndpointFlag := flag.String("first", ":9100", "first host to scrape for metrics diff")
	rhsEndpointFlag := flag.String("second", ":9100", "second host to scrape for metrics diff")
	deviationFlag := flag.Float64("deviation", 0, "deviation amount to consider noteworthy")
	truncateLenFlag := flag.Int("len", 0, "allowed length of timeseries column. 0 allows all (default)")

	flag.Parse()

	lhsResults := make(map[string]float64)
	rhsResults := make(map[string]float64)

	// TODO(mfuller): I really don't like this, but idk how to improve it yet.
	//  The Prometheus client model nests what we _really_ want (series+labels, vals) deep down
	//  and in a way not easily extracted
	// TODO(mfuller): There's an alternative approach here; take the MetricFamily, marshal it to text (prometheus/expfmt
	//  can do this), and then iterate each line. _Maaaaybe_ that's more efficient?
	lhsJob := scrape.NewScrapeJob(*lhsEndpointFlag, types.NewTinderboxHTTPOptions())
	for r := range scrape.OpenMetricScrape(lhsJob, *http.DefaultClient) {
		if r.Error() != nil {
			log.Fatalln(r.Error())
		}

		for fName, fam := range r.Families() {
			for _, m := range fam.Metric {
				mavN, mavV := extractMetricAndVal(fName, fam.GetType(), m)
				lhsResults[mavN] = mavV
			}
		}
	}
	rhsJob := scrape.NewScrapeJob(*rhsEndpointFlag, types.NewTinderboxHTTPOptions())
	for r := range scrape.OpenMetricScrape(rhsJob, *http.DefaultClient) {
		if r.Error() != nil {
			log.Fatalln(r.Error())
		}

		for fName, fam := range r.Families() {
			for _, m := range fam.Metric {
				mavN, mavV := extractMetricAndVal(fName, fam.GetType(), m)
				rhsResults[mavN] = mavV
			}
		}
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	for _, notableSeries := range wrap(lhsResults, rhsResults, *deviationFlag) {
		writeOut(w, notableSeries, *truncateLenFlag)
	}
	w.Flush()
}

func writeOut(w io.Writer, result HandedResult, len int) {
	// For series where deviation is !NaN, write with a little pizzaz
	if !math.IsNaN(result.deviation) {
		color.New(color.Bold, color.FgHiMagenta).Fprintln(w, result.AsTabbed(len))
		return
	}
	color.New(color.FgCyan).Fprintln(w, result.AsTabbed(len))
}

type handedness int

func (h handedness) String() string {
	switch h {
	case LeftHanded:
		return "<"
	case RightHanded:
		return ">"
	case BothHanded:
		return "|"
	}
	return "?"
}

const (
	LeftHanded handedness = iota
	RightHanded
	BothHanded
)

const (
	SeriesNotFound = " "
)

// HandedResult allows for easier processing of a left-handed and right-handed {timeseries, value} pair, as well as
// convenience fields for signifying noteworthy data (noteworthy side, value absolute deviation)
type HandedResult struct {
	notableSide handedness // representing which side (or both) is notable for this HandedResult
	lhName      string     // left-side series name+labels
	rhName      string     // right-side series name+labels
	lhVal       float64    // left-side value
	rhVal       float64    // right-side value
	deviation   float64    // deviation between lhVal and rhVal
}

// AsTabbed returns a tabwriter-ready string representation of this HandedResult
func (h HandedResult) AsTabbed(len int) string {
	return fmt.Sprintf("%s\t%f\t%s\t%s\t%f\t%f",
		truncateTo(h.lhName, len),
		h.lhVal,
		h.notableSide,
		truncateTo(h.rhName, len),
		h.rhVal,
		h.deviation,
	)
}

// wrap takes a left-and-right side map[string]float64 representing timeseries and values from two Prometheus/OTel endpoints.
// It also takes a deviation for calling out spicy differences, and then returns a unified slice of HandedResult
func wrap(l, r map[string]float64, deviation float64) []HandedResult {
	var ret []HandedResult
	for series, lhV := range l {
		rhV, ok := r[series]
		if !ok {
			// Unique to l
			ret = append(ret, HandedResult{
				notableSide: LeftHanded,
				lhName:      series,
				rhName:      SeriesNotFound,
				lhVal:       lhV,
				rhVal:       math.NaN(),
				deviation:   math.NaN(),
			})
			continue
		}
		// Found in r, so determine if vals deviate enough
		absDiff := math.Abs(rhV) - math.Abs(lhV)
		if absDiff >= deviation {
			// and if they do...
			ret = append(ret, HandedResult{
				notableSide: BothHanded,
				lhName:      series,
				rhName:      series,
				lhVal:       lhV,
				rhVal:       rhV,
				deviation:   absDiff,
			})
		}
		delete(r, series)
	}
	// Add anything still in r
	for series, rhV := range r {
		ret = append(ret, HandedResult{
			notableSide: RightHanded,
			lhName:      SeriesNotFound,
			rhName:      series,
			lhVal:       math.NaN(),
			rhVal:       rhV,
			deviation:   math.NaN(),
		})
	}
	return ret
}

// truncateTo returns a truncated version of s if len(s) > allowed length. If smaller than the allowed length (or allowed
// length is zero), the entire string is returned
func truncateTo(s string, allowedLen int) string {
	runeStr := []rune(s)
	if len(runeStr) <= allowedLen || allowedLen == 0 {
		return s
	}

	return string(runeStr[0:allowedLen]) + "..."
}

// consistentCollapseLabels takes a LabelPair and collapses it into a sorted "prometheus-like" string
// such as: a=valA,b=valB,...
func consistentCollapseLabels(l []*dto.LabelPair) string {
	var s []string
	for _, v := range l {
		collapsed := fmt.Sprintf(`%s="%s"`, v.GetName(), v.GetValue())
		s = append(s, collapsed)
	}
	sort.Strings(s)
	return strings.Join(s, ",")
}

// getValue takes a Metric and MetricType, and attempts to return the value. Summaries and Histograms are pains in my
// ass and not _as_ interesting for what spod is meant to highlight, so getValue returns values from Counter and Gauge
// metrics only
func getValue(from *dto.Metric, t dto.MetricType) float64 {
	switch t {
	case dto.MetricType_COUNTER:
		return from.Counter.GetValue()
	case dto.MetricType_GAUGE:
		return from.Gauge.GetValue()
	default:
		// TODO(mfuller): this..
	}
	return -1
}

// extractMetricAndVal takes a series name (e.g. node_cpu_seconds_total), a metric type, and the actual metric to process.
// It then builds a consistent, sorted "fake" label set, gets the metric value, and returns those two beautiful pieces
// of data.
func extractMetricAndVal(name string, t dto.MetricType, m *dto.Metric) (string, float64) {
	ccls := consistentCollapseLabels(m.GetLabel())

	if len(ccls) == 0 {
		return name, getValue(m, t)
	}

	return fmt.Sprintf(`%s{%s}`, name, ccls), getValue(m, t)
}
