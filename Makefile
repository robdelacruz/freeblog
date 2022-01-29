# Usage:
# 'make dep' and 'make webtools' to install dependencies.
# 'make clean' to clear all work files
# 'make' to build css and js into static/
# 'make serve' to start dev webserver

NODE_VER = 14

JSFILES = index.js helpers.js Dashboard.svelte Entries.svelte EditEntry.svelte DelEntry.svelte Images.svelte EditImage.svelte DelImage.svelte Files.svelte EditFile.svelte DelFile.svelte AccountMenu.svelte EditSite.svelte EditUserSettings.svelte ChangePassword.svelte DelUser.svelte UploadImages.svelte SearchImages.svelte FileThumbnail.svelte FileLink.svelte PopupMenu.svelte Tablinks.svelte

all: freeblog static/style.css static/bundle.js

dep:
	curl -fsSL https://deb.nodesource.com/setup_$(NODE_VER).x | sudo bash -
	sudo apt install nodejs
	sudo npm install -g npx
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
	rm -rf freeblog static/bundle.js static/*.css static/*.map

serve:
	python -m SimpleHTTPServer

