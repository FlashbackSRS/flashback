package repo

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/flimzy/testify/require"

	"github.com/FlashbackSRS/flashback-model"
	"github.com/FlashbackSRS/flashback/cardmodel/mock"
)

func TestPrepareBody(t *testing.T) {
	mock.RegisterMock("mock-model")
	require := require.New(t)
	doc := strings.NewReader(testDoc1)
	result, err := prepareBody(Question, 0, "mock-model", doc)
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
<script type="text/javascript">
		/* Mock Model */
		console.log("Mock Model 'mock-model'");
</script></head>
<body class="card">
    Question: <img src="paste-13877039333377.jpg"/><br/><div><sub>instrument</sub></div>
</body></html>`

type PrioTest struct {
	Due      fb.Due
	Interval fb.Interval
	Expected float64
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
}

func TestPrio(t *testing.T) {
	now := parseTime("2017-01-01 00:00:00")
	for _, test := range PrioTests {
		prio := CardPrio(&test.Due, &test.Interval, now)
		if math.Abs(float64(prio)-test.Expected) > 0.000001 {
			t.Errorf("%s / %s: Expected priority %f, got %f\n", test.Due, test.Interval, test.Expected, prio)
		}
	}
}

func parseTime(ts string) time.Time {
	t, err := time.Parse("2006-01-02 15:04:05", ts)
	if err != nil {
		panic(err)
	}
	return t
}

func parseDue(ds string) fb.Due {
	d, err := fb.ParseDue(ds)
	if err != nil {
		panic(err)
	}
	return d
}
