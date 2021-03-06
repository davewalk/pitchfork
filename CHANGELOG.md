# Change Log
All notable changes to this project will be documented in this file. It follows [Semantic Versioning](http://semver.org) guidelines. 

## 0.3.1 - 2015-02-11
### Fixed
- Multi-word artist names were not working with the search command. This was fixed by taking all command line arguments besides the first and concatenating them into a string to search on.  

## 0.3.0 - 2015-02-11
### Added
- "search <artist name>" command added for searching past reviews.  

## 0.2.2 - 2015-02-02
### Fixed
- An album's score was not being returned because there were multiple ".score" elements on the review page. This is fixed by only taking the text of the first matched element.

## 0.2.1 - 2015-02-02
### Fixed
- Returning the year of the album as a string since reissues could include the initial release year and the reissue year ("1976/2015").
- Improved error handling in retrieving review details overall.

## 0.2.0 - 2015-02-01
### Added
- "news" command  added for returning latest Pitchfork news (up to the 10 latest articles)
- Alternate flags for most flags (see `pitchfork help` for details)
- Separated elements for extracting data from Pitchfork.com into "pitchfork" package

## 0.1.0 - 2015-01-25
### Added
- Initial version of the app released.
