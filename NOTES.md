# Notes

## Database

### Format
- Table `files`
  - Column `file_id` primary key
  - Column `storage_request` JSON
  - Column `time`: datetime of last update
  - Column for each provider
    - Either null or a string representing access information
- Table `queue`
  - Column `rowid`
  - Column `file_id`
  - Column `provider`
  - Column `status`: `0 pending`, `1 in-progress`, `2 success`, `3 failed`
  - Column `taken` bool - if a thread is currently working on the task or not
  - Column `time`: datetime of last update

### Usage
1. New storage request comes in
2. It's added to `files`, with all providers set to null
3. Each provider is added to the queue as `0 pending` and not taken


## Config
- `AB_DATA_DIR` env var

## Todo
- More DB funcs
- Logging
- Config
