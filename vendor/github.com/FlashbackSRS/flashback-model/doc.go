/*
Package fb defines the basic data structures used by the FlashbackSRS project.

This model is designed with the assumption that these data structures are stored
using a CouchDB/PouchDB backend. Additionally, the Package struct allows for
convenient packaging of multiple objects in a single JSON structure, which may
then be exported and imported as a stand-alone file.

The primary object types are used as follows:

Package

Intended for the import and export of JSON structures containing an entire
collection of documents. Roughly synonymous with Anki's *.apkg format.

Bundle

A bundle represents a distinct PouchDB/CouchDB database. This is where most of
the other document types (Cards being the exception), live. Cards live in a
UserDB (to be documented elsewhere. TODO: Doc this).

Theme

A theme is a collection of Models.

Model

A model is essentially a card type definition. It defines the layout, including
fields, of a card type. A model must have one or more templates.

Template

A template is what defines the front and back of a given card.

Note

A note contains the data necessary to populate a model. Notes are the core of
studyable material.

Card

A card is the logical combination of a Note and a Model View, and represents the
review element to be studied. Card documents are stored in the UserDB, as they
are unique per user,  whereas all other doc types may potentially be shared
among multiple users.

*/
package fb
