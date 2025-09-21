# Test Scripts

## Usage
```bash
# Run in sequence:
./test-auth-endpoints.sh      # Must run first
./test-files-endpoints.sh     # Run after auth
# Set FILE_PATH inside the file
# or just execute it like this: FILE_PATH=/you/file/path.png ./test-files-endpoints.sh
./test-product-endpoints.sh   # Run after auth
./test-profile-endpoints.sh   # Run after auth
./test-purchase-endpoints.sh  # Run after auth
```

## Dependencies
- **Auth** → Generates `test-data-auth-*` (used by all others except purchase)
- **Files** → Generates `test-data-file` (used by profile, product, purchase)
- **Profile** → No output, TODO: add file_id in payload
- **Product** → Should generate `test-data-product` (used by purchase), TODO: use file_id from test-data-file
- **Purchase** → No output, TODO: use product_id from test-data-product, file_id from test-data-file

## Current Limitations
- Profile doesn't use file_id yet
- Product uses hardcoded file_id
- Purchase uses hardcoded product_id and file_id

## File Path Setup
- Update `FILE_PATH` in `test-files-endpoints.sh` to point to your test image file.
- You can also override it by running the script like this: `FILE_PATH=/you/file/path.png ./test-files-endpoints.sh`