'use strict';
docReady(function() {
    var face = FB.face == 0 ? 'question' : 'answer';
    if ( face == 'answer' && FB.card.context !== undefined ) {
        console.log(FB.card);
        var answers = FB.card.context.typedAnswers;
        if ( answers !== undefined ) {
            for ( var key in answers ) {
                if ( answers.hasOwnProperty(key) ) {
                    var field = document.getElementsByName('type:'+key)[0]; // There should be only one
                    if ( field !== undefined ) {
                        field.value = answers[key].text;
                        // FIXME: Do something with answers[key].correct (i.e. disable the 'correct' buttons?)
                        field.setAttribute('disabled', true);
                    } else {
                        console.log("Got an answer for an unknown field: " + key);
                    }
                }
            }
        }
    }
});
