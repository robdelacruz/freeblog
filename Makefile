# Usage:
# 'make dep' and 'make webtools' to install dependencies.
# 'make clean' to clear all work files
# 'make' to build css and js into static/
# 'make serve' to start dev webserver

JSFILES = index.js helpers.js Dashboard.svelte Entries.svelte EditEntry.svelte DelEntry.svelte Images.svelte EditImage.svelte DelImage.svelte UploadImages.svelte SearchImages.svelte FileThumbnail.svelte PopupMenu.svelte Tablinks.svelte

all: freeblog static/style.css static/bundle.js

dep:
	sudo apt update
	sudo apt install curl software-properties-common
	curl -sL https://deb.nodesource.com/setup_13.x | sudo bash -
	sudo apt install nodejs
	sudo npm --force install -g npx
	go get github.com/gorilla/feeds
	go get github.com/shurcooL/github_flavored_markdown

webtools:
	npm install --save-dev tailwindcss
	npm install --save-dev postcss-cli
	npm install --save-dev cssnano
	npm install --save-dev svelte
	npm install --save-dev rollup
	npm install --save-dev rollup-plugin-svelte
	npm install --save-dev @rollup/plugin-node-resolve

static/style.css: twsrc.css
	#npx tailwind build twsrc.css -o twsrc.o 1>/dev/null
	#npx postcss twsrc.o > static/style.css
	npx tailwind build twsrc.css -o static/style.css 1>/dev/null

static/bundle.js: $(JSFILES)
	npx rollup -c

freeblog: freeblog.go
	go build -o freeblog freeblog.go

clean:
	rm -rf freeblog static/*.js static/*.css static/*.map

serve:
	python -m SimpleHTTPServer

