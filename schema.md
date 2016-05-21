Bundle Database
===============

                ID convention
Bundle          Normal                                  Anki IMport                                 example
------
id          -- SHA1(Owner's UUID + 64 random bits)      SHA1(Owner's UUID + "-anki-" + col.crt)     cfea19a055d32aa1b0133544d5ec4f63b2cd3779
name
description
created
modified
imported
owner
permissions* -- Undetermined

Theme
-----
id          -- BundleID+":"+64 random bits              Bundle ID + ":" + 64-bit model.ID   cfea19a055d32aa1b0133544d5ec4f63b2cd3779:043ea39
parent      -- May be null.References a parent Theme, from which this was cloned, and from which changes may be auto-incorporated.
bundle_id
name
description
created
modified
imported

Model
-----
id          -- ThemeID + ":" + 8-bit counter            ThemeID + ":" + 8-bit counter       cfea19a055d32aa1b0133544d5ec4f63b2cd3779:043ea39:0
type    (Standard, Cloze, etc)
name
description
created
modified
imported
fields[]
attachments[]

Note
----
id          -- BundleID + ":" + 64 random bits          BundleID + ":" + 64-bit note.ID     cfea19a055d32aa1b0133544d5ec4f63b2cd3779:043ea39:0
parent      -- May be null. References a parent note, from which this one was cloned, and from which changes may be auto-incorporated.
model_id    -- May reference a model in any bundle
created
modified
imported
values[]
attachments[]
tags[]
instances[] -- Computed list of card instances which are output when the note is passed through the model (strings)

Card -- Implicit
----
id          -- NoteID + ":" + ModelID + ":" + instance_id (string?)

Deck
----
id          -- BundleID + ":" + 64 random bits          BundleID + ":" + deck.ID
name
description
created
modified
imported
cards[]     -- May reference cards in other bundles
decks[]     -- May (probably will) reference decks in other bundles


User Database
=============

Bundle Stub
-----------
id      BundleID


CardStat
--------
card_id
last_studied_date
suspended   (bool? list of reasons?)
tags[]
notes[]     (bug reports, etc)
etc, etc
