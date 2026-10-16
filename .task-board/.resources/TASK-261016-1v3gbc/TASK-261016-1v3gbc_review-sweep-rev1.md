# Review sweep
Candidate e1515c85c5ed4e3dbf932fb4549bf7eaea66e621. Four mandatory paired rows used because no catalog-keyed table was supplied.

1. broken — closed metadata: earlier candidate cases and null-channel refusal passed, but the later TestReviewerGreekCollision admits four forbidden file/component pairs; finding unicode-casefold-renormalization. Logs focused.log, controls.log, row1.log.

2. held — complete actual tree: metadata-only directory refused by Verify and PortableEqual; candidate modes/links/FIFO/read-error attacks pass. Logs focused.log, row2.log.

3. held — portable identity: left-side drift refuses after a valid comparison; candidate source/version/profile/descriptor/file/mode pairs pass. Logs focused.log, row3.log.

4. held — baseline retention: producer-version recapture refuses twice against the original digest; candidate changed-content recapture and immutable-copy assertions pass. Logs focused.log, row4.log.

Free hunt: launch hardlink, sticky content mode and absent launch metadata all refused; four valid Unicode path spellings accepted. Superseded by the additional Greek free-hunt result: unicode-casefold-renormalization, bypass, four variants accepted. See greek-sensitive.log. Logs free-hunt.log, unicode.log.
