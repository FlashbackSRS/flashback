HTML_FILES = $(shell find webclient/html/ -type f -name '*.html')
CSS_FILES = $(shell find webclient/css/ -type f -name '*.css')
PNG_FILES = $(shell find webclient/images/ -type f -name '*.png')
SVG_FILES = $(shell find webclient/images/ -type f -name '*.svg')
WEBP_FILES = $(shell find webclient/images/ -type f -name '*.webp')
I18N_FILES = $(wildcard translations/*.all.json)

COUCH_SERVER = $(shell echo $$FLASHBACK_COUCH_URL | sed -e 's|//.*@|//|')

# I know this is an ugly hack, but it works on my dev machine. If ever
# anyone else is using this, we can find a better solution
#FLASHBACK_API_BASEURI = https://$(shell ip -f inet addr show wlan0 | tr -s ' ' | egrep '^\s*inet' | cut -d' ' -f3 | cut -d'/' -f1):4002/

server: www
	go run ./server/*

plugins: www .phonegap-facebook-plugin
	mkdir -p plugins
	cordova plugin add https://git-wip-us.apache.org/repos/asf/cordova-plugin-console.git
	cordova plugin add .phonegap-facebook-plugin --variable APP_ID=$(FLASHBACK_FACEBOOK_ID) --variable APP_NAME="$(FLASHBACK_FACEBOOK_NAME)"
#	cordova plugin add https://git-wip-us.apache.org/repos/asf/cordova-plugin-device.git

.phonegap-facebook-plugin:
	git clone https://github.com/Wizcorp/phonegap-facebook-plugin.git .phonegap-facebook-plugin

platforms: www
	mkdir -p platforms
	cordova platform | grep -q Installed.*android || cordova platform add android

cordova-init: plugins platforms

android: cordova-init cordova-www
	cordova run android

go-test:
	go test
	gopherjs test
	gopherjs test github.com/flimzy/flashback/webclient/pages/all/
# 	go test

test: go-test

clean:
	rm -rf www bundle.js pre-bundle.js bundle.js.map flashback main.js main.js.map

distclean:
	rm -rf node_modules plugins platforms

npm-install: package.json
	npm install

javascript: www/js/flashback.js www/js/worker.sql.js www/js/cardframe.js
www/js/flashback.js: package.json webclient/main.js main.js npm-install
	mkdir -p www/js
#	browserify js/main.js > $@ || ( stats=$?; rm -f $@; exit $? )
	browserify --debug . -o pre-bundle.js
	cat pre-bundle.js | exorcist bundle.js.map > bundle.js
	cp pre-bundle.js $@
# 	uglifyjs bundle.js -c -m -o $@ \
# 		--source-map $@.map \
# 		--source-map-root /js \
# 		--source-map-url /js/flashback.js.map \
# 		--in-source-map bundle.js.map

www/js/worker.sql.js: webclient/vendor/sql.js/worker.sql.js
	cp $< $@

www/js/cardframe.js: webclient/js/cardframe.js
	cp $< $@

.PHONY: main.js
main.js:
	gopherjs build ./webclient/*.go
# 	uglifyjs main.js -c -m -o $@

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
www: javascript css images $(HTML_FILES) $(I18N_FILES)
	mkdir -p www/translations
	cp $(HTML_FILES) www
	cp $(I18N_FILES) www/translations
	sed -i -e 's|__API_SERVER__|$(FLASHBACK_API_BASEURI)|g' www/index.html
	sed -i -e 's|__COUCH_SERVER__|$(COUCH_SERVER)|g' www/index.html

cordova-www: www
	cat www/index.html | sed -e 's/<!-- Cordova Here -->/<script src="cordova.js"><\/script>/' > www/cordova.html
