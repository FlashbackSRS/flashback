Layout of a Theme
=================

Theme
-----
A theme is a collection of related (probably visually) card models.

Model
-----
A card model contains a master template in the format of Go's [html/template](https://golang.org/pkg/html/template/) package. This template may include an arbitrary number of other sub-templates.

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
