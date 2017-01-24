package repo

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/FlashbackSRS/flashback-model"
	"github.com/FlashbackSRS/flashback/webclient/views/studyview"
	"github.com/flimzy/testify/require"
)

// We need to implement our own, minimal fake controller here, because using
// controllers/mock results in an import cycle.
type fakeController struct{}

func (f *fakeController) Type() string                               { return "fake-model" }
func (f *fakeController) IframeScript() []byte                       { return []byte("/* Fake Model */") }
func (f *fakeController) Buttons(_ int) (studyview.ButtonMap, error) { return nil, nil }
func (f *fakeController) Action(_ *PouchCard, _ *int, _ time.Time, _ studyview.Button) (bool, error) {
	return false, nil
}

func TestPrepareBody(t *testing.T) {
	require := require.New(t)
	doc := strings.NewReader(testDoc1)
	result, err := prepareBody(Question, 0, &fakeController{}, doc)
	if err != nil {
		t.Errorf("error preparing body: %s", err)
	}
	require.HTMLEqual(expected1, result, "prepareBody did something funky")
}

var testDoc1 = `<!DOCTYPE html>
<html><head>
<title>FB Card</title>
<base href="https://flashback.ddns.net:4001/">
<meta charset="UTF-8">
<meta http-equiv="Content-Security-Policy" content="script-src 'unsafe-inline' https://flashback.ddns.net:4001/">
<script type="text/javascript">
'use strict';
var FB = {
iframeID: '445a737462464b4e',
};
</script>
<script type="text/javascript" src="js/cardframe.js"></script>
<script type="text/javascript"></script>
<style></style>
</head>
<body class="card">

<div class="question" data-id="0">
    Question: <img src="paste-13877039333377.jpg"><br><div><sub>instrument</sub></div>
</div>
<div class="answer" data-id="0">
    Question: <img src="paste-13877039333377.jpg"><br><div><sub>instrument</sub></div>

<hr id="answer">

Answer: <div>instrumento</div><div>[sound:pronunciation_es_instrumento.3gp]</div>
</div>

<div class="question" data-id="1">
    Question: <div>instrumento</div><div>[sound:pronunciation_es_instrumento.3gp]</div>
</div>
<div class="answer" data-id="1">
    <hr id="answer">

<br>
Answer: <img src="paste-13877039333377.jpg"><br><div><sub>instrument</sub></div>
</div>


</body></html>
    `

var expected1 = `<!DOCTYPE html><html><head>
<title>FB Card</title>
<base href="https://flashback.ddns.net:4001/"/>
<meta charset="UTF-8"/>
<meta http-equiv="Content-Security-Policy" content="script-src &#39;unsafe-inline&#39; https://flashback.ddns.net:4001/"/>
<script type="text/javascript">
'use strict';
var FB = {
iframeID: '445a737462464b4e',
};
</script>
<script type="text/javascript" src="js/cardframe.js"></script>
<script type="text/javascript"></script>
<style></style>
<script type="text/javascript">/* Fake Model */</script></head>
<body class="card">
    Question: <img src="paste-13877039333377.jpg"/><br/><div><sub>instrument</sub></div>
</body></html>`

type PrioTest struct {
	Due      fb.Due
	Interval fb.Interval
	Expected float64
	Now      time.Time
}

var PrioTests = []PrioTest{
	PrioTest{
		Due:      parseDue("2017-01-01 00:00:00"),
		Interval: fb.Day,
		Expected: 1,
	},
	PrioTest{
		Due:      parseDue("2017-01-01 12:00:00"),
		Interval: fb.Day,
		Expected: 0.125,
	},
	PrioTest{
		Due:      parseDue("2016-12-31 12:00:00"),
		Interval: fb.Day,
		Expected: 3.375,
	},
	PrioTest{
		Due:      parseDue("2017-02-01 00:00:00"),
		Interval: 60 * fb.Day,
		Expected: 0.112912,
	},
	PrioTest{
		Due:      parseDue("2017-01-02 00:00:00"),
		Interval: fb.Day,
		Expected: 0,
	},
	PrioTest{
		Due:      parseDue("2016-01-02 00:00:00"),
		Interval: 7 * fb.Day,
		Expected: 150084.109375,
	},
	PrioTest{
		Due:      parseDue("2017-01-24 11:16:59"),
		Interval: 10 * fb.Minute,
		Expected: 132.520996,
		Now:      parseTime("2017-01-24T11:57:58+01:00"),
	},
}

func TestPrio(t *testing.T) {
	for _, test := range PrioTests {
		if test.Now.IsZero() {
			test.Now = parseTime("2017-01-01 00:00:00")
		}
		prio := CardPrio(&test.Due, &test.Interval, test.Now)
		if math.Abs(float64(prio)-test.Expected) > 0.000001 {
			t.Errorf("%s / %s: Expected priority %f, got %f\n", test.Due, test.Interval, test.Expected, prio)
		}
	}
}

var timeFormats = []string{
	time.RFC3339,
	"2006-01-02 15:04:05",
}

func parseTime(ts string) time.Time {
	var err error
	var t time.Time
	for _, format := range timeFormats {
		t, err = time.Parse(format, ts)
		if err == nil {
			return t
		}
	}
	panic(fmt.Sprintf("invalid time value '%s'", ts))
}

func parseDue(ds string) fb.Due {
	d, err := fb.ParseDue(ds)
	if err != nil {
		panic(err)
	}
	return d
}

func TestUnmarshal(t *testing.T) {
	raw := `{"_id":"card-alnlcvykyjxsjtijzonc3456kd5u4757.udROb8T8RmRASG5zGHNKnKL25zI.0","_rev":"1-daccd83780014e8cf35ce8f16d2a144c","created":"2015-09-08T23:55:03.000000539Z","imported":"2017-01-02T17:16:56.764985035+01:00","model":"theme-ELr8cEJJOvJU4lYz-VTXhH8wLTo/0","modified":"2016-08-02T13:05:04Z","type":"card"}`
	card := &PouchCard{}
	if err := json.Unmarshal([]byte(raw), card); err != nil {
		t.Errorf("Failed to unmarshal card: %s\n", err)
	}
}
