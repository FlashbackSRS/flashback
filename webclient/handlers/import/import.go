package import_handler

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/flimzy/flashback-model"
	"github.com/flimzy/flashback/repository"
	"github.com/flimzy/goweb/file"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
)

var jQuery = jquery.NewJQuery

func BeforeTransition(event *jquery.Event, ui *js.Object, p url.Values) bool {
	go func() {
		container := jQuery(":mobile-pagecontainer")
		jQuery("#importnow", container).On("click", func() {
			fmt.Printf("Attempting to import something...\n")
			go func() {
				if err := DoImport(); err != nil {
					fmt.Printf("Error importing: %s\n", err)
				}
				fmt.Printf("DoImport() returned\n")
			}()
		})
		jQuery(".show-until-load", container).Hide()
		jQuery(".hide-until-load", container).Show()
	}()

	return true
}

func DoImport() error {
	files := file.InternalizeFileList(jQuery("#apkg", ":mobile-pagecontainer").Get(0).Get("files"))
	for i := 0; i < files.Length; i++ {
		if err := importFile(files.Item(i)); err != nil {
			return err
		}
	}
	fmt.Printf("Done with import\n")
	return nil
}

func importFile(f *file.File) error {
	u, err := repo.CurrentUser()
	if err != nil {
		return err
	}
	fmt.Printf("Gonna pretend to import %s now\n", f.Name)
	b, err := f.Bytes()
	if err != nil {
		return err
	}
	z, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(z)
	z.Close()
	pkg := &fb.Package{}
	if err := json.Unmarshal(buf.Bytes(), pkg); err != nil {
		return err
	}

	bundle := pkg.Bundle
	udb, err := u.DB()
	if err != nil {
		return err
	}
	if err := udb.Save(bundle); err != nil {
		return err
	}
	bundle.Rev = nil
	bdb, err := repo.BundleDB(bundle)
	if err != nil {
		return err
	}
	if err := bdb.Save(bundle); err != nil {
		return err
	}

	// From this point on, we plow through the errors
	errs := make([]error, 0)

	// Themes
	for _, t := range pkg.Themes {
		if err := bdb.Save(t); err != nil {
			fmt.Printf("Error saving theme: %s\n", err)
			errs = append(errs, err)
		}
	}
	// Notes
	for _, n := range pkg.Notes {
		if err := bdb.Save(n); err != nil {
			fmt.Printf("Error saving note: %s\n", err)
			errs = append(errs, err)
		}
	}
	// Decks
	for _, d := range pkg.Decks {
		if err := bdb.Save(d); err != nil {
			fmt.Printf("Error saving deck: %s\n", err)
			errs = append(errs, err)
		}
	}
	// Cards
	// 	for _, c := range pkg.Cards {
	// 		fmt.Printf("Saving card: %v\n", c)
	// 		if err := udb.Save(c); err != nil {
	// 			fmt.Printf("Error saving card: %s\n", err)
	// 			errs = append(errs, err)
	// 		}
	// 	}

	// Did we have errors?
	if len(errs) > 0 {
		for i, e := range errs {
			fmt.Printf("Error %d: %s\n", i, e)
		}
		return fmt.Errorf("%d errors encountered", len(errs))
	}

	return nil
}
