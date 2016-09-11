package test

import (
	"regexp"
	"testing"

	"github.com/flimzy/testify/require"

	"github.com/FlashbackSRS/flashback/repository"
)

var revRE = regexp.MustCompile(`"_rev":"\d-[0-9a-f]+?"`)
var iframeRE = regexp.MustCompile(`iframeID: '[0-9a-f]+?'`)

func TestCard1(t *testing.T) {
	require := require.New(t)

	testImport(t)

	u := repo.User{testUser}

	card, err := u.GetCard("card-iw5x7ie66fsepm67hey2fqjms6fywi6v.4WpHslICjKMtkmw-KKpSJECrnuc.0")
	require.Nil(err, "Error fetching card: %s", err)
	require.NotNil(card, "Card is nil")

	body, _, err := card.Body()
	require.Nil(err, "Unable to fetch the card's body: %s", err)
	body = revRE.ReplaceAllString(body, `"_rev":"X-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"`)
	body = iframeRE.ReplaceAllString(body, `iframeID: 'xxxxxxxxxxxxxxxx'`)
	require.LinesEqual(expectedBody1, body, "Card's body")
}

func TestCard2(t *testing.T) {
	require := require.New(t)

	testImport(t)

	u := repo.User{testUser}

	card, err := u.GetCard("card-iw5x7ie66fsepm67hey2fqjms6fywi6v.4WpHslICjKMtkmw-KKpSJECrnuc.1")
	require.Nil(err, "Error fetching card: %s", err)
	require.NotNil(card, "Card is nil")

	body, _, err := card.Body()
	require.Nil(err, "Unable to fetch the card's body: %s", err)
	body = revRE.ReplaceAllString(body, `"_rev":"X-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"`)
	body = iframeRE.ReplaceAllString(body, `iframeID: 'xxxxxxxxxxxxxxxx'`)
	require.LinesEqual(expectedBody2, body, "Card's body")
}

var expectedBody1 = `<!DOCTYPE html><html><head>
	<title>FB Card</title>
	<base href=""/>
	<meta charset="UTF-8"/>
	<meta http-equiv="Content-Security-Policy" content="script-src &#39;unsafe-inline&#39; "/>
<script type="text/javascript">
'use strict';
var FB = {
	iframeID: 'xxxxxxxxxxxxxxxx',
	card: {"type":"card","_id":"card-iw5x7ie66fsepm67hey2fqjms6fywi6v.4WpHslICjKMtkmw-KKpSJECrnuc.0","_rev":"X-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx","created":"0001-01-01T00:00:00Z","modified":"0001-01-01T00:00:00Z"},
	note: {"type":"note","_id":"note-4WpHslICjKMtkmw-KKpSJECrnuc","_rev":"X-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx","created":"2015-09-09T00:00:40.000000521Z","modified":"2016-08-02T13:15:00Z","imported":"2016-09-11T16:48:32.699714842+02:00","theme":"theme-94hk99pCpQ5DAMGZvpb5_HR5oqs","model":0,"fieldValues":[{"text":"\u003cimg src=\"paste-14413910245377.jpg\" /\u003e\u003cbr /\u003e\u003cdiv\u003e\u003csub\u003eart\u003c/sub\u003e\u003c/div\u003e","files":["paste-14413910245377.jpg"]},{"text":"\u003cdiv\u003earte\u003c/div\u003e\u003cdiv\u003e[sound:pronunciation_es_arte.3gp]\u003c/div\u003e","files":["pronunciation_es_arte.3gp"]}],"_attachments":{"paste-14413910245377.jpg":{"content_type":"image/jpeg","data":"/9j/4AAQSkZJRgABAQEAkACQAAD/2wBDAP//////////////////////////////////////////////////////////////////////////////////////2wBDAf//////////////////////////////////////////////////////////////////////////////////////wAARCAC4ARIDASIAAhEBAxEB/8QAFwABAQEBAAAAAAAAAAAAAAAAAAECA//EACAQAQEBAAIDAQEBAQEAAAAAAAABESExAhJRYUFxgaH/xAAWAQEBAQAAAAAAAAAAAAAAAAAAAQL/xAAZEQEBAQADAAAAAAAAAAAAAAAAAREhQWH/2gAMAwEAAhEDEQA/ANBiigIAAoAAAAAAAAIVIAuavCaguSJaiUDbVkn9I0IZAUEXDpm0GuGbU5q4Cc1qGs3y5BsTU1FVNE3ABNFHQAQBN0IAKoAAAAAgAAIAICdgpiyNAkilsjNtvQjWs3yMq5ICZaZDWbQa1nWTAXf4YuLxAF4jO/EFW3TE05ojWQZwB1Bm0IADSgoiAAAAAAAAjQoM4uCWiGyJzU9vw9gX1+rsjGoDV8mdQAMakXASRrJE1L/oLvxE2Q20DZE5q41gM41i5Iuip6hoDTLRRIyJyDTQyoYohoigAAaAuCaC9M6ImmFusgoIqCAsjUgM41ipUU1OanRqocRNtWTWsBnGsVQTFvCWs7opbojUmiINes+gNAAJigJiNAusmNAanSf8aBGGuAFEVEVAQCstJipRZGpAAQZFQDBMXF1GkOlZtNBrZE1kAO1z6bgLJIXyY07Bfah60B0ASpATk/T1VVO1UAAEFBEarKKCaCiLjWKms41IuAhWHRAYXGqzqYonYm1UXpO1wTVTFBNEw2Rb05rEW1Gp4tSKMzx+tyLigmCgIAIgomCwZ3KbvSrOWhz39XaGNib+LAGcbQGcXFABQEVEtwFS34x7HYL/AOgqKmKggAACBgLISNY0iY0KCKAAmwBAFQRaiDNqANrMXWQHWXU3Kx01LvYmNDM4rQigAiW4tc6LC2stLIDMjeLiiM9IVGVURQQXFxcExcIqhFARUEBdc75WljOgvImgOqoCKlGLdFgi6UaRf8RueIJP1cSwnwF/poCNKgIGCgmJVrIrQAhiYoCYYKKhQAVlQAEBFZAvLM8WlkUZ9YNgigCIzY0CysLI0ousYsqVNFa1lFBYupLidg1pKlJ9GW1ZlUAw2AIqVIK0ACWgAAAipQFQRBUGpASRoFQAAiKlAAEAZ2gvkyu7EGoADQqNCM1qQxVZNS1cYqLFnbbMaCiYoICAoAAUSgAUVCGNSIiSNAqAJvIKGwBE0FSgACWACQBGwAKKAhysvACFrPYC9LqAI1PJQFNTABQBEAFMUAUAQABLcTsATAFR/9k="},"pronunciation_es_arte.3gp":{"content_type":"audio/3gpp","data":"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"}}},
	model: {"id":0,"modelType":0,"name":"Basic-24b78","templates":["Card 1","Card 2"],"fields":[{"fieldType":3,"name":"Front"},{"fieldType":3,"name":"Back"}],"files":["!Basic-24b78.Card 1 answer.html","!Basic-24b78.Card 1 question.html","!Basic-24b78.Card 2 answer.html","!Basic-24b78.Card 2 question.html","$template.0.html"]},
};
</script>
<script type="text/javascript" src="js/cardframe.js"></script>
<script type="text/javascript"></script>
<style></style>
</head>
<body class="card">
		<img src="data:image/jpeg;base64,/9j/4AAQSkZJRgABAQEAkACQAAD/2wBDAP//////////////////////////////////////////////////////////////////////////////////////2wBDAf//////////////////////////////////////////////////////////////////////////////////////wAARCAC4ARIDASIAAhEBAxEB/8QAFwABAQEBAAAAAAAAAAAAAAAAAAECA//EACAQAQEBAAIDAQEBAQEAAAAAAAABESExAhJRYUFxgaH/xAAWAQEBAQAAAAAAAAAAAAAAAAAAAQL/xAAZEQEBAQADAAAAAAAAAAAAAAAAAREhQWH/2gAMAwEAAhEDEQA/ANBiigIAAoAAAAAAAAIVIAuavCaguSJaiUDbVkn9I0IZAUEXDpm0GuGbU5q4Cc1qGs3y5BsTU1FVNE3ABNFHQAQBN0IAKoAAAAAgAAIAICdgpiyNAkilsjNtvQjWs3yMq5ICZaZDWbQa1nWTAXf4YuLxAF4jO/EFW3TE05ojWQZwB1Bm0IADSgoiAAAAAAAAjQoM4uCWiGyJzU9vw9gX1+rsjGoDV8mdQAMakXASRrJE1L/oLvxE2Q20DZE5q41gM41i5Iuip6hoDTLRRIyJyDTQyoYohoigAAaAuCaC9M6ImmFusgoIqCAsjUgM41ipUU1OanRqocRNtWTWsBnGsVQTFvCWs7opbojUmiINes+gNAAJigJiNAusmNAanSf8aBGGuAFEVEVAQCstJipRZGpAAQZFQDBMXF1GkOlZtNBrZE1kAO1z6bgLJIXyY07Bfah60B0ASpATk/T1VVO1UAAEFBEarKKCaCiLjWKms41IuAhWHRAYXGqzqYonYm1UXpO1wTVTFBNEw2Rb05rEW1Gp4tSKMzx+tyLigmCgIAIgomCwZ3KbvSrOWhz39XaGNib+LAGcbQGcXFABQEVEtwFS34x7HYL/AOgqKmKggAACBgLISNY0iY0KCKAAmwBAFQRaiDNqANrMXWQHWXU3Kx01LvYmNDM4rQigAiW4tc6LC2stLIDMjeLiiM9IVGVURQQXFxcExcIqhFARUEBdc75WljOgvImgOqoCKlGLdFgi6UaRf8RueIJP1cSwnwF/poCNKgIGCgmJVrIrQAhiYoCYYKKhQAVlQAEBFZAvLM8WlkUZ9YNgigCIzY0CysLI0ousYsqVNFa1lFBYupLidg1pKlJ9GW1ZlUAw2AIqVIK0ACWgAAAipQFQRBUGpASRoFQAAiKlAAEAZ2gvkyu7EGoADQqNCM1qQxVZNS1cYqLFnbbMaCiYoICAoAAUSgAUVCGNSIiSNAqAJvIKGwBE0FSgACWACQBGwAKKAhysvACFrPYC9LqAI1PJQFNTABQBEAFMUAUAQABLcTsATAFR/9k="/><br/><div><sub>art</sub></div>
	</body></html>`

var expectedBody2 = `<!DOCTYPE html><html><head>
	<title>FB Card</title>
	<base href=""/>
	<meta charset="UTF-8"/>
	<meta http-equiv="Content-Security-Policy" content="script-src &#39;unsafe-inline&#39; "/>
<script type="text/javascript">
'use strict';
var FB = {
	iframeID: 'xxxxxxxxxxxxxxxx',
	card: {"type":"card","_id":"card-iw5x7ie66fsepm67hey2fqjms6fywi6v.4WpHslICjKMtkmw-KKpSJECrnuc.1","_rev":"X-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx","created":"0001-01-01T00:00:00Z","modified":"0001-01-01T00:00:00Z"},
	note: {"type":"note","_id":"note-4WpHslICjKMtkmw-KKpSJECrnuc","_rev":"X-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx","created":"2015-09-09T00:00:40.000000521Z","modified":"2016-08-02T13:15:00Z","imported":"2016-09-11T16:48:32.699714842+02:00","theme":"theme-94hk99pCpQ5DAMGZvpb5_HR5oqs","model":0,"fieldValues":[{"text":"\u003cimg src=\"paste-14413910245377.jpg\" /\u003e\u003cbr /\u003e\u003cdiv\u003e\u003csub\u003eart\u003c/sub\u003e\u003c/div\u003e","files":["paste-14413910245377.jpg"]},{"text":"\u003cdiv\u003earte\u003c/div\u003e\u003cdiv\u003e[sound:pronunciation_es_arte.3gp]\u003c/div\u003e","files":["pronunciation_es_arte.3gp"]}],"_attachments":{"paste-14413910245377.jpg":{"content_type":"image/jpeg","data":"/9j/4AAQSkZJRgABAQEAkACQAAD/2wBDAP//////////////////////////////////////////////////////////////////////////////////////2wBDAf//////////////////////////////////////////////////////////////////////////////////////wAARCAC4ARIDASIAAhEBAxEB/8QAFwABAQEBAAAAAAAAAAAAAAAAAAECA//EACAQAQEBAAIDAQEBAQEAAAAAAAABESExAhJRYUFxgaH/xAAWAQEBAQAAAAAAAAAAAAAAAAAAAQL/xAAZEQEBAQADAAAAAAAAAAAAAAAAAREhQWH/2gAMAwEAAhEDEQA/ANBiigIAAoAAAAAAAAIVIAuavCaguSJaiUDbVkn9I0IZAUEXDpm0GuGbU5q4Cc1qGs3y5BsTU1FVNE3ABNFHQAQBN0IAKoAAAAAgAAIAICdgpiyNAkilsjNtvQjWs3yMq5ICZaZDWbQa1nWTAXf4YuLxAF4jO/EFW3TE05ojWQZwB1Bm0IADSgoiAAAAAAAAjQoM4uCWiGyJzU9vw9gX1+rsjGoDV8mdQAMakXASRrJE1L/oLvxE2Q20DZE5q41gM41i5Iuip6hoDTLRRIyJyDTQyoYohoigAAaAuCaC9M6ImmFusgoIqCAsjUgM41ipUU1OanRqocRNtWTWsBnGsVQTFvCWs7opbojUmiINes+gNAAJigJiNAusmNAanSf8aBGGuAFEVEVAQCstJipRZGpAAQZFQDBMXF1GkOlZtNBrZE1kAO1z6bgLJIXyY07Bfah60B0ASpATk/T1VVO1UAAEFBEarKKCaCiLjWKms41IuAhWHRAYXGqzqYonYm1UXpO1wTVTFBNEw2Rb05rEW1Gp4tSKMzx+tyLigmCgIAIgomCwZ3KbvSrOWhz39XaGNib+LAGcbQGcXFABQEVEtwFS34x7HYL/AOgqKmKggAACBgLISNY0iY0KCKAAmwBAFQRaiDNqANrMXWQHWXU3Kx01LvYmNDM4rQigAiW4tc6LC2stLIDMjeLiiM9IVGVURQQXFxcExcIqhFARUEBdc75WljOgvImgOqoCKlGLdFgi6UaRf8RueIJP1cSwnwF/poCNKgIGCgmJVrIrQAhiYoCYYKKhQAVlQAEBFZAvLM8WlkUZ9YNgigCIzY0CysLI0ousYsqVNFa1lFBYupLidg1pKlJ9GW1ZlUAw2AIqVIK0ACWgAAAipQFQRBUGpASRoFQAAiKlAAEAZ2gvkyu7EGoADQqNCM1qQxVZNS1cYqLFnbbMaCiYoICAoAAUSgAUVCGNSIiSNAqAJvIKGwBE0FSgACWACQBGwAKKAhysvACFrPYC9LqAI1PJQFNTABQBEAFMUAUAQABLcTsATAFR/9k="},"pronunciation_es_arte.3gp":{"content_type":"audio/3gpp","data":"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"}}},
	model: {"id":0,"modelType":0,"name":"Basic-24b78","templates":["Card 1","Card 2"],"fields":[{"fieldType":3,"name":"Front"},{"fieldType":3,"name":"Back"}],"files":["!Basic-24b78.Card 1 answer.html","!Basic-24b78.Card 1 question.html","!Basic-24b78.Card 2 answer.html","!Basic-24b78.Card 2 question.html","$template.0.html"]},
};
</script>
<script type="text/javascript" src="js/cardframe.js"></script>
<script type="text/javascript"></script>
<style></style>
</head>
<body class="card">
		<div>arte</div><div>[sound:pronunciation_es_arte.3gp]</div>
	</body></html>`
