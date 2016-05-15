Layout of a Theme
=================

Theme
-----
A theme is a collection of related (probably visually) card models.

Theme ID = hex encoded({owner UUID} + {int64 of Anki Model} + {1 byte Anki model type})

Model
-----
A card model contains a master template in the format of Go's [html/template](https://golang.org/pkg/html/template/) package. This template may include an arbitrary number of other sub-templates.

ModelID = ThemeID + '/' + Model # (0 indexed)

Note
----
A note represents the data presented on cards. Each note has a single parent Model.
Each note may be in one or more Decks.

Deck
----
A deck has no parent object type. It contains zero or more Notes. It may contain Notes of different types and models.

+------------------------------------------------------------------------------+
| `Theme`                                                                      |
|  Attachments: template.1.html, template.2.html, script.js, style.css, etc    |
|                                                                              |
|  +----------------------------------+ +----------------------------------+   |
|  | `Model`                          | | `Model`                          |   |
|  |                                  | |                                  |   |
|  |  +---------------------------+   | |  +----------------------------+  |   |
|  |  | template.1.html           |   | |  | template.2.html            |  |   |
|  |  |                           |   | |  |                            |  |   |
