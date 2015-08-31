package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gorilla/handlers"
	"net/http"
)

func main() {
	if httpsErr := BindHTTPS(); httpsErr != nil {
		log.Print(httpsErr)
		if httpErr := BindHTTP(); httpErr != nil {
			log.Print(httpErr)
		}
	} else {
		if httpErr := RedirectHTTP(); httpErr != nil {
			log.Print(httpErr)
		}
	}
	select {}
}

func BindHTTPS() error {
	bindAddress := os.Getenv("FLASHBACK_HTTPS_BIND")
	if len(bindAddress) == 0 {
		return fmt.Errorf("FLASHBACK_HTTPS_BIND notset, not serving HTTPS")
	}
	log.Printf("Serving static assets via HTTPS on %s\n", bindAddress)
	go func() {
		log.Fatal(http.ListenAndServeTLS(bindAddress, os.Getenv("FLASHBACK_SSL_CERT"), os.Getenv("FLASHBACK_SSL_KEY"),
			handlers.LoggingHandler(os.Stderr, http.FileServer(http.Dir("www")))))
	}()
	return nil
}

func BindHTTP() error {
	bindAddress := os.Getenv("FLASHBACK_HTTP_BIND")
	if len(bindAddress) == 0 {
		return fmt.Errorf("FLASHBACK_HTTP_BIND not set, not serving HTTP")
	}
	log.Printf("Redirecting for static assets via HTTP on %s\n\n", bindAddress)
	go func() {
		log.Fatal(http.ListenAndServe(bindAddress, handlers.LoggingHandler(os.Stderr, http.FileServer(http.Dir("www")))))
	}()
	return nil
}

func RedirectHTTP() error {
	bindAddress := os.Getenv("FLASHBACK_HTTP_BIND")
	if len(bindAddress) == 0 {
		return fmt.Errorf("FLASHBACK_HTTP_BIND not set, not redirecting HTTP")
	}
	log.Printf("Serving static assets via HTTP on %s\n", bindAddress)
	go func() {
		log.Fatal(http.ListenAndServe(bindAddress, handlers.LoggingHandler(os.Stderr, http.HandlerFunc(RedirectHandler))))
	}()
	return nil
}

func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	baseURI := os.Getenv("FLASHBACK_BASEURI")
	if len(baseURI) == 0 {
		log.Fatal("FLASHBACK_BASEURI must be set if FLASHBACK_HTTP_BIND is set\n")
	}
	// Redirect all HTTP requests to HTTPS
	// It is the responsibility of the admin to configure FLASHBACK_API_BASEURI
	// properly.  I will not be held responsible for broken redirections
	// or redirect loops!
	http.Redirect(w, r, baseURI, http.StatusFound)
}
