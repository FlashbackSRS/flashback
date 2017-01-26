'use strict';
var requests = {};
var iframeID;
window.addEventListener('error', function(e) {
    if (typeof(iframeID) === 'undefined') {
        var metas = document.getElementsByTagName('meta');
        for (var i=0; i < metas.length; i++) {
            if (metas[i].getAttribute("name") == "iframeid") {
                iframeID = metas[i].getAttribute("content");
                break;
            }
        }
        if (typeof iframeID === 'undefined') {
            console.log("Didn't find the iframe ID!!");
        }
    }
    var t = e.target;
    var tag = t.tagName;
    var path = tag == 'LINK' ? t.getAttribute('href') : t.getAttribute('src');
    if ( typeof requests[tag] === 'undefined' ) {
        requests[tag] = {};
    }
    if ( typeof requests[tag][path] === 'undefined' ) {
        requests[tag][path] = {
            targets: []
        };
        parent.postMessage({
            IframeID: iframeID,
            Tag: tag,
            CardID: FB.card.id,
            Path: path,
        }, '*');
    }
    if ( typeof requests[tag][path].url === 'undefined' ) {
        // This means we already requested the file, but it hasn't arrived yet
        requests[tag][path].targets.push( t );
        return false;
    }
    // We've already requested the file, and we have it now!
    t.src = requests[tag][path].url;
    return false;
}, true);

window.addEventListener('message', function(e) {
    var tag = e.data.Tag,
        path = e.data.Path,
        type = e.data.ContentType,
        data = e.data.Data;
    var b = new Blob([data], {type: type})
    var url = URL.createObjectURL(b)

    var req = requests[ tag ][ path ];
    req.url = url
    var attr = data.Tag == 'LINK' ? 'href' : 'src';
    for ( var i = 0; i < req.targets.length; i++ ) {
        req.targets[i][attr] = url;
    }
    req.targets = [];
});
