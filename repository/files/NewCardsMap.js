function(doc) {
    if ( doc._id.lastIndexOf('card-', 0) !== 0 )    return;
    if ( doc.due !== undefined )                    return;
    if ( doc.suspended === true )                   return;
    emit( doc.created, { created: doc.created, buriedUntil: doc.buriedUntil } );
}