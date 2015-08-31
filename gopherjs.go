// +build js

package flashback

import (
    "github.com/gopherjs/gopherjs/js"
    "github.com/gopherjs/jquery"
    "honnef.co/go/js/console"
    
    "sync"
)

//var jQuery = jquery.NewJQuery

func Console(args ...interface{}) {
    console.Log(args...)
}

func (c *FlashbackClient) Get(uri string) (string,error) {
    var wg sync.WaitGroup
    
    var document string
    var err error
    
    wg.Add(1)
    jquery.Ajax(map[string]interface{} {
        "url": uri,
        "dataType": "html",
    }).Done(func (d string, s string, x *js.Object) {
        defer wg.Done()
        document = d
    }).Fail(func (x *js.Object, s string, e *js.Error) {
        defer wg.Done()
        err = e
    })
    wg.Wait()
    return document,err
}
