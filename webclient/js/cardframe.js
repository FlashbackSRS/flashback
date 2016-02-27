'use strict';
var requests = {};
window.addEventListener('error', function(e) {
    var t = e.target;
    var tag = t.tagName;
    var path = tag == 'LINK' ? t.getAttribute('href') : t.getAttribute('src');
//     console.log("not found : " + path);
    if ( typeof requests[tag] === 'undefined' ) {
        requests[tag] = {};
    }
    if ( typeof requests[tag][path] === 'undefined' ) {
        requests[tag][path] = {
            targets: []
        };
        parent.postMessage(JSON.stringify({
            IframeId: FB.iframeId,
            Tag: tag,
            CardId: FB.card._id,
            NoteId: FB.note._id,
            ModelId: FB.model._id,
            Path: path,
        }), '*');
    }
    if ( typeof requests[tag][path].data === 'undefined' ) {
        // This means we already requested the file, but it hasn't arrived yet
        requests[tag][path].targets.push( t );
        return false;
    }
    // We've already requested the file, and we have it now!
    t.src = requests[tag][path].data;
    return false;
}, true);

window.addEventListener('message', function(e) {
//     console.log(e);
    var data = e.data;
// console.log("attempting to activate " + data.Tag + " / " + data.Path);
    var req = requests[ data.Tag ][ data.Path ];
    req.data = data.Data;
    var attr = data.Tag == 'LINK' ? 'href' : 'src';
    for ( var i = 0; i < req.targets.length; i++ ) {
        req.targets[i][attr] = data.Data;
    }
    req.targets = [];
});
