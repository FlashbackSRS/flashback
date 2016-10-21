'use strict';

// answer submits an answer to the parent, authorizing the parent to close this
// page
function answer(a) {
    parent.postMessage(JSON.stringify({
        IframeID: FB.iframeID,
        Tag: tag,
        CardId: FB.card._id,
        NoteId: FB.note._id,
        ModelId: FB.model._id,
        Answer: a
    }), '*');
}
