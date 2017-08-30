// Package log provides a simple wrapper around the standard Go library's 'log'
// package, combining concepts from two of Dave Cheney's blog posts:
//
//  - http://dave.cheney.net/2014/09/28/using-build-to-switch-between-debug-and-release
//  - http://dave.cheney.net/2015/11/05/lets-talk-about-logging
//
// In particular, it exposes only the log.Print* functions, and adds an
// additional set of log.Debug* functions, which in turn call their log.Print*
// counterparts only when the 'debug' build tag is present.
package log
