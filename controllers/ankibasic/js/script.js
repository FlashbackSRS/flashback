'use strict';
docReady(function() {
    var face = FB.face == 0 ? 'question' : 'answer';
    if ( face == 'answer' && FB.card.context !== undefined ) {
        console.log(FB.card);
        var answers = FB.card.context.typedAnswers;
        if ( answers !== undefined ) {
            for ( var key in answers ) {
                if ( answers.hasOwnProperty(key) && key.lastIndexOf('type:', 0) == 0 ) {
                    var field = document.getElementsByName(key)[0]; // There should be only one
                    if ( field !== undefined ) {
                        field.value = answers[key];
                        field.setAttribute('disabled', true);
                    }
                }
            }
        }
    }
});
