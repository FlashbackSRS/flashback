The Flashback API is intended to be as simple as possible, to allow for the greatest interoperability, while providing the necessary features and security.

The primary data store for Flashback is stored in [CouchDB](http://couchdb.apache.org/), and simply uses the CouchDB API.  However, certain conventions are used within this framework, and they are documented here.

## CouchDB Document conventions

CouchDB provides a few requirements for every document, and naturally Flashback follows those conventions.  Additionally, Flashback follows a few other conventions outlined here.

### Key names

- Keys names use PascalCase for values which are expected to be used by clients, and in the Go implementations, are direclty mapped to struct field names.  Examples: `FullName`, `Email`

- Key names begin with a '$' character and use camelCase for values intended for internal use, such ass uuids and foreign keys, and timestamps, except _id, and other CouchDB-required keys

- Key names use camelCase for values which are not expected to be used by clients directly.  Some clients may use such values, but probably after a data transform of some sort. You ought to consult the relevant portion of this document before using one of these values, to understand how it is to be used.

- If you are extending any of these documents outside of the official project, you are encouraged to prefix your key names with 'X-' or 'x-'. These values are reserved for custom implementations, and will never be used by the official project.
