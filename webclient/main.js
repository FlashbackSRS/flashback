window.$ = require('jquery');
window.PouchDB = require('pouchdb');
window.PouchDB.plugin( require('pouchdb-find') );
// jQuery Mobile has to be initialized *after* the mobileinit event handler
// is configured in the GopherJS code, so we put the jQM initialization in
// this special function which GopherJS will call at the appropriate time.
loadjqueryMobile = function() {
    window.$.mobile = require('jquery-mobile');
};
require('main'); // GopherJS compiled code
