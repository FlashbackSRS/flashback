'use strict';

/* This function is borrowed from http://stackoverflow.com/a/9899701/13860 */
(function(funcName, baseObj) {
    // The public function name defaults to window.docReady
    // but you can pass in your own object and own function name and those will be used
    // if you want to put them in a different namespace
    funcName = funcName || "docReady";
    baseObj = baseObj || window;
    var readyList = [];
    var readyFired = false;
    var readyEventHandlersInstalled = false;

    // call this when the document is ready
    // this function protects itself against being called more than once
    function ready() {
        if (!readyFired) {
            // this must be set to true before we start calling callbacks
            readyFired = true;
            for (var i = 0; i < readyList.length; i++) {
                // if a callback here happens to add new ready handlers,
                // the docReady() function will see that it already fired
                // and will schedule the callback to run right after
                // this event loop finishes so all handlers will still execute
                // in order and no new ones will be added to the readyList
                // while we are processing the list
                readyList[i].fn.call(window, readyList[i].ctx);
            }
            // allow any closures held by these functions to free
            readyList = [];
        }
    }

    function readyStateChange() {
        if ( document.readyState === "complete" ) {
            ready();
        }
    }

    // This is the one public interface
    // docReady(fn, context);
    // the context argument is optional - if present, it will be passed
    // as an argument to the callback
    baseObj[funcName] = function(callback, context) {
        // if ready has already fired, then just schedule the callback
        // to fire asynchronously, but right away
        if (readyFired) {
            setTimeout(function() {callback(context);}, 1);
            return;
        } else {
            // add the function and context to the list
            readyList.push({fn: callback, ctx: context});
        }
        // if document already ready to go, schedule the ready function to run
        if (document.readyState === "complete") {
            setTimeout(ready, 1);
        } else if (!readyEventHandlersInstalled) {
            // otherwise if we don't have event handlers installed, install them
            if (document.addEventListener) {
                // first choice is DOMContentLoaded event
                document.addEventListener("DOMContentLoaded", ready, false);
                // backup is window load event
                window.addEventListener("load", ready, false);
            } else {
                // must be IE
                document.attachEvent("onreadystatechange", readyStateChange);
                window.attachEvent("onload", ready);
            }
            readyEventHandlersInstalled = true;
        }
    }
})("docReady", window);

(function(baseObj) {
    function fserve(r) {
        var path = r.path,
            type = r.content_type,
            data = r.data;
        var b = new Blob([data], {type: type})
        var url = URL.createObjectURL(b)

        var req = requests[ path ];
        var attr = data.Tag == 'LINK' ? 'href' : 'src';
        for ( var i = 0; i < req.targets.length; i++ ) {
            var target = req.targets[i];
            var attr = 'src' in target ? 'src' : 'href';
            target[attr] = url;
        }
        req.targets = [];
    }

    var requests = {};
    baseObj.addEventListener('error', function(e) {
        var t = e.target;
        if ( t === undefined ) {
            // This isn't a 404 error; so just skip it
            return true;
        }
        var path = t.tagName == 'LINK' ? t.getAttribute('href') : t.getAttribute('src');
        if ( requests[path] === undefined ) {
            requests[path] = {
                targets: []
            };
            parent.postMessage({
                type:     "fserve",
                iframeID: window.location.href,
                payload:  path,
            }, '*');
        }
        if ( requests[path].url === undefined ) {
            // This means we already requested the file, but it hasn't arrived yet
            requests[path].targets.push( t );
            return false;
        }
        // We've already requested the file, and we have it now!
        t.src = requests[path].url;
        return false;
    }, true);

    baseObj.addEventListener('message', function(e) {
        switch (e.data.type) {
            case "fserve":
                fserve(e.data.payload);
                break;
            default:
                console.log("Unexpected message type '" + e.data.type + "' from parent")
        }
    });
})(window);

function playaudio() {
    var media = document.getElementsByTagName("audio")
    if (media.length == 0) {
        return;
    }
    console.log(media[0].src);
    // Install 'ended' event handlers on all but the last media file, which
    // will trigger the following one to play.
    for (var i=0; i < media.length-1; i++) {
        media[i].addEventListener('ended', function() {
            playWhenReady(media[i+1]);
        });
    }
    playWhenReady(media[0]);
}

function playWhenReady(media) {
    // A lazy way to check if the media file has been loaded yet.
    if ( media.src.match(/^blob:/) ) {
        media.play();
        return;
    }
    media.addEventListener('canplay', function() {
        media.play();
    });
}

docReady(playaudio);
