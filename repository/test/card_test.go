package test

import (
	"regexp"
	"testing"

	"github.com/flimzy/testify/require"

	"github.com/FlashbackSRS/flashback/controllers/mock"
	"github.com/FlashbackSRS/flashback/repository"
)

var revRE = regexp.MustCompile(`"_rev":"\d-[0-9a-f]+?"`)
var iframeRE = regexp.MustCompile(`iframeID: '[0-9a-f]+?'`)

func init() {
	mock.RegisterMock("anki-basic")
}

func TestCard1(t *testing.T) {
	require := require.New(t)

	testImport(t)

	u := repo.User{testUser}

	card, err := u.GetCard("card-alnlcvykyjxsjtijzonc3456kd5u4757.ZR4TpeX38xRzRvXprlgJpP4Ribo.0")
	require.Nil(err, "Error fetching card: %s", err)
	require.NotNil(card, "Card is nil")

	question, _, err := card.Body(repo.Question)
	require.Nil(err, "Unable to fetch the card's question body: %s", err)
	question = revRE.ReplaceAllString(question, `"_rev":"X-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"`)
	question = iframeRE.ReplaceAllString(question, `iframeID: 'xxxxxxxxxxxxxxxx'`)
	require.LinesEqual(expectedQuestion0, question, "Card 0 question")

	answer, _, err := card.Body(repo.Answer)
	require.Nil(err, "Unable to fetch the card's answer body: %s", err)
	answer = revRE.ReplaceAllString(answer, `"_rev":"X-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"`)
	answer = iframeRE.ReplaceAllString(answer, `iframeID: 'xxxxxxxxxxxxxxxx'`)
	require.LinesEqual(expectedAnswer0, answer, "Card 0 answer")
}

func TestCard2(t *testing.T) {
	require := require.New(t)

	testImport(t)

	u := repo.User{testUser}

	card, err := u.GetCard("card-alnlcvykyjxsjtijzonc3456kd5u4757.ZR4TpeX38xRzRvXprlgJpP4Ribo.1")
	require.Nil(err, "Error fetching card: %s", err)
	require.NotNil(card, "Card is nil")

	question, _, err := card.Body(repo.Question)
	require.Nil(err, "Unable to fetch the card's body: %s", err)
	question = revRE.ReplaceAllString(question, `"_rev":"X-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"`)
	question = iframeRE.ReplaceAllString(question, `iframeID: 'xxxxxxxxxxxxxxxx'`)
	require.LinesEqual(expectedQuestion1, question, "Card 1 question")

	answer, _, err := card.Body(repo.Answer)
	require.Nil(err, "Unable to fetch the card's answer body: %s", err)
	answer = revRE.ReplaceAllString(answer, `"_rev":"X-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"`)
	answer = iframeRE.ReplaceAllString(answer, `iframeID: 'xxxxxxxxxxxxxxxx'`)
	require.LinesEqual(expectedAnswer1, answer, "Card 1 answer")
}

var expectedQuestion0 = `<!DOCTYPE html><html><head>
	<title>FB Card</title>
	<base href=""/>
	<meta charset="UTF-8"/>
	<meta http-equiv="Content-Security-Policy" content="script-src &#39;unsafe-inline&#39; "/>
<script type="text/javascript">
'use strict';
var FB = {
	iframeID: 'xxxxxxxxxxxxxxxx',
	card: {"id":"card-alnlcvykyjxsjtijzonc3456kd5u4757.ZR4TpeX38xRzRvXprlgJpP4Ribo.0"},
	note: {"id":"note-ZR4TpeX38xRzRvXprlgJpP4Ribo"}
};
</script>
<script type="text/javascript" src="js/cardframe.js"></script>
<script type="text/javascript"></script>
<style></style>
<script type="text/javascript">
		/* Mock Model */
		console.log("Mock Model 'anki-basic'");
</script></head>
<body class="card">
		Question: <img src="paste-13877039333377.jpg"/><br/><div><sub>instrument</sub></div>
	</body></html>`

var expectedAnswer0 = `<!DOCTYPE html><html><head>
	<title>FB Card</title>
	<base href=""/>
	<meta charset="UTF-8"/>
	<meta http-equiv="Content-Security-Policy" content="script-src &#39;unsafe-inline&#39; "/>
<script type="text/javascript">
'use strict';
var FB = {
	iframeID: 'xxxxxxxxxxxxxxxx',
	card: {"id":"card-alnlcvykyjxsjtijzonc3456kd5u4757.ZR4TpeX38xRzRvXprlgJpP4Ribo.0"},
	note: {"id":"note-ZR4TpeX38xRzRvXprlgJpP4Ribo"}
};
</script>
<script type="text/javascript" src="js/cardframe.js"></script>
<script type="text/javascript"></script>
<style></style>
<script type="text/javascript">
		/* Mock Model */
		console.log("Mock Model 'anki-basic'");
</script></head>
<body class="card">
		Question: <img src="paste-13877039333377.jpg"/><br/><div><sub>instrument</sub></div>

<hr id="answer"/>

Answer: <div>instrumento</div><div>[sound:pronunciation_es_instrumento.3gp]</div>
	</body></html>`

var expectedQuestion1 = `<!DOCTYPE html><html><head>
	<title>FB Card</title>
	<base href=""/>
	<meta charset="UTF-8"/>
	<meta http-equiv="Content-Security-Policy" content="script-src &#39;unsafe-inline&#39; "/>
<script type="text/javascript">
'use strict';
var FB = {
	iframeID: 'xxxxxxxxxxxxxxxx',
	card: {"id":"card-alnlcvykyjxsjtijzonc3456kd5u4757.ZR4TpeX38xRzRvXprlgJpP4Ribo.1"},
	note: {"id":"note-ZR4TpeX38xRzRvXprlgJpP4Ribo"}
};
</script>
<script type="text/javascript" src="js/cardframe.js"></script>
<script type="text/javascript"></script>
<style></style>
<script type="text/javascript">
		/* Mock Model */
		console.log("Mock Model 'anki-basic'");
</script></head>
<body class="card">
		Question: <div>instrumento</div><div>[sound:pronunciation_es_instrumento.3gp]</div>
	</body></html>`

var expectedAnswer1 = `<!DOCTYPE html><html><head>
	<title>FB Card</title>
	<base href=""/>
	<meta charset="UTF-8"/>
	<meta http-equiv="Content-Security-Policy" content="script-src &#39;unsafe-inline&#39; "/>
<script type="text/javascript">
'use strict';
var FB = {
	iframeID: 'xxxxxxxxxxxxxxxx',
	card: {"id":"card-alnlcvykyjxsjtijzonc3456kd5u4757.ZR4TpeX38xRzRvXprlgJpP4Ribo.1"},
	note: {"id":"note-ZR4TpeX38xRzRvXprlgJpP4Ribo"}
};
</script>
<script type="text/javascript" src="js/cardframe.js"></script>
<script type="text/javascript"></script>
<style></style>
<script type="text/javascript">
		/* Mock Model */
		console.log("Mock Model 'anki-basic'");
</script></head>
<body class="card">
		<hr id="answer"/>

<br/>
Answer: <img src="paste-13877039333377.jpg"/><br/><div><sub>instrument</sub></div>
	</body></html>`
