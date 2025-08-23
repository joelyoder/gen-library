# gen-library
Image organizer for local generation

## Backend configuration

The backend logger can be configured with environment variables:

| Variable   | Description                                            |
|------------|--------------------------------------------------------|
| `LOG_LEVEL` | Sets the log verbosity. Accepts `debug`, `info` (default), `warn`, or `error`. |
| `LOG_FILE`  | When set, writes logs to `logs/$(LOG_FILE)` in addition to stdout. The previous log file is rotated to `$(LOG_FILE).1`. |

To enable file logging, set `LOG_FILE` to a file name. For example, `LOG_FILE=backend.log` will log to `logs/backend.log` as well as standard output.

## Todo

### Metadata
- Link to Model Manager to show model names as links that open the model in MM

### Image viewer
- Allow for zoom (and pinch and zoom) in the image viewer
- Nicer metadata display with buttons to copy prompts with one click

### Settings
- Change NSFW keyword list to setting

### Gallery
- Change NSFW button to indicator that can toggle NSFW status on the fly

### Misc
- Add icon instead of app name
- Add tests
