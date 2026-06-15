ASSET_FILES := LICENSE README.md fix_avatar_colors_for_overlay font schedules static switch_config.txt templates tunnel

all:
	@rm -rf r7-arena/
	@go clean
	@mkdir r7-arena
	@go build -o r7-arena/
	@cp -r $(ASSET_FILES) r7-arena/
	@chmod +x ./r7-arena/r7-arena
	@./r7-arena/r7-arena