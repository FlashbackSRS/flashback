[![GoDoc](https://godoc.org/github.com/flimzy/log?status.png)](http://godoc.org/github.com/flimzy/log)

Package `log` provides a simple wrapper around the standard Go library's `log`
package, combining concepts from two of Dave Cheney's blog posts:

 - [Using // +build to switch between debug and release builds](http://dave.cheney.net/2014/09/28/using-build-to-switch-between-debug-and-release)
 - [Let’s talk about logging](http://dave.cheney.net/2015/11/05/lets-talk-about-logging)

In particular, it exposes only the `log.Print*` functions, and adds an
additional set of `log.Debug*` functions, which in turn call their `log.Print*`
counterparts only when the 'debug' build tag is present.

This package is released under the terms of the MIT license. See the included
LICENSE.txt for details.
