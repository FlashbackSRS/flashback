package repo

import (
	"strings"
	"testing"

	"github.com/FlashbackSRS/flashback/cardmodel"
	"github.com/FlashbackSRS/flashback/cardmodel/mock"
	"github.com/flimzy/testify/require"
)

func TestPrepareBody(t *testing.T) {
	cardmodel.RegisterModel(&mock.Model{}) // Register the mock handler
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
        /* Placeholder JS */
        console.log("Mock Handler");
    </script></head>
<body class="card">
    Question: <img src="paste-13877039333377.jpg"/><br/><div><sub>instrument</sub></div>
</body></html>`
