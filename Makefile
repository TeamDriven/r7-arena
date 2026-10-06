ASSET_FILES := LICENSE README.md fix_avatar_colors_for_overlay font schedules static switch_config.txt templates tunnel event.db

all:
	sudo /sbin/rc-service r7-arena stop
	rm -f event.db
	mkdir -p r7-arena/
	test -f r7-arena/event.db && cp r7-arena/event.db ./
	rm -rf r7-arena/
	go clean
	mkdir r7-arena
	go build -o r7-arena/
	cp -r $(ASSET_FILES) r7-arena/
	chmod +x ./r7-arena/r7-arena
	sudo /sbin/rc-service r7-arena start
