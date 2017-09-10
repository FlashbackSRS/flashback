HTML_FILES = $(shell find webclient/html/ -type f -name '*.html')
CSS_FILES = $(shell find webclient/css/ -type f -name '*.css')
PNG_FILES = $(shell find webclient/images/ -type f -name '*.png')
SVG_FILES = $(shell find webclient/images/ -type f -name '*.svg')
WEBP_FILES = $(shell find webclient/images/ -type f -name '*.webp')
I18N_FILES = $(wildcard translations/*.all.json)

# Borrowed from https://stackoverflow.com/questions/9551416/gnu-make-how-to-join-list-and-separate-it-with-separator
# A literal space.
space :=
space +=

# Joins elements of the list in arg 2 with the given separator.
#   1. Element separator.
#   2. The list.
join-with = $(subst $(space),$1,$(strip $2))

all: www android

android: GOTAGS = cordova
android: cordova-www
	cordova prepare
	cordova run android

go-test: preclean npm-install generate
	gopherjs test $$(go list ./... | grep -v /vendor/)

test: go-test

preclean:
	rm -rf ${GOPATH}/pkg/*_js

clean:
	rm -rf www bundle.js pre-bundle.js bundle.js.map flashback main.js main.js.map

distclean:
	rm -rf node_modules plugins platforms

npm-install: package.json
	npm install

javascript: www/js/flashback.js www/js/worker.sql.js www/js/cardframe.js
www/js/flashback.js: package.json webclient/main.js main.js npm-install
	mkdir -p www/js
ifdef FLASHBACK_PROD
	browserify --exclude pouchdb-all-dbs --exclude xhr2 . -o bundle.js
	uglifyjs bundle.js -m -o $@ --stats
	# uglifyjs bundle.js -c -m -o $@ --stats
	# \
	# 	--stats \
	# 	--source-map $@.map \
	# 	--source-map-root /js \
	# 	--source-map-url /js/flashback.js.map \
	# 	--in-source-map bundle.js.map
else
	browserify --exclude pouchdb-all-dbs --exclude xhr2 --debug . -o pre-bundle.js
	cp pre-bundle.js $@
endif

www/js/worker.sql.js: webclient/vendor/sql.js/worker.sql.js
	cp $< $@

www/js/cardframe.js: webclient/js/cardframe.js
	cp $< $@

.PHONY: main.js
main.js: preclean generate
ifdef FLASHBACK_PROD
	gopherjs build --tags="$(call join-with,$(space),$(GOTAGS))" -m ./webclient -o main.js
	# uglifyjs main.js -c -m -o $@
else
	gopherjs build --tags="$(call join-with,$(space),$(GOTAGS) debug)" ./webclient -o main.js
endif

css: webclient/vendor/jquery.mobile-1.4.5/jquery.mobile.inline-svg-1.4.5.min.css webclient/vendor/jquery.mobile-1.4.5/images/ajax-loader.gif $(CSS_FILES)
	mkdir -p www/css/images
	cp webclient/vendor/jquery.mobile-1.4.5/jquery.mobile.inline-svg-1.4.5.min.css www/css/jquery.mobile-1.4.5.css
	cp webclient/vendor/jquery.mobile-1.4.5/images/ajax-loader.gif www/css/images
#	yui-compressor vendor/jquery.mobile-1.4.5/jquery.mobile-1.4.5.css -o www/css/jquery.mobile-1.4.5.css
#	cp -a vendor/jquery.mobile-1.4.5/images www/css
	cp webclient/css/*.css www/css

images: $(PNG_FILES) $(WEBP_FILES) $(SVG_FILES) webclient/images/favicon.ico
	mkdir -p www/images
	cp $(PNG_FILES) www/images
	cp $(WEBP_FILES) www/images
	cp $(SVG_FILES) www/images
	cp webclient/images/favicon.ico www

.PHONY: www
www: javascript css images $(HTML_FILES) $(I18N_FILES) generate check-env
	mkdir -p www/translations
	cp $(HTML_FILES) www
	cp $(I18N_FILES) www/translations
	sed -i -e 's|__API_SERVER__|$(FLASHBACK_BASEURI)|g' www/index.html
	sed -i -e 's|__FACEBOOK_ID__|$(FLASHBACK_FACEBOOK_ID)|g' www/index.html

check-env:
ifndef FLASHBACK_BASEURI
    $(error FLASHBACK_BASEURI is undefined)
endif
ifndef FLASHBACK_FACEBOOK_ID
    $(error FLASHBACK_FACEBOOK_ID is undefined)
endif

cordova-www: www
	cat www/index.html | sed -e 's/<!-- Cordova Here -->/<script src="cordova.js"><\/script>/' > www/cordova.html

generate:
	go generate $$(go list ./... | grep -v /vendor/)
