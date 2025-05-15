### 0.4.0
*Compatible with outputs from Ripple1D Pipeline version 0.10.3 to present*

- A new command `domain` has been added to create a composite domain map for the given reaches.
- Format of extent library has been modified, this is to improve performance of `fim` command. This changes API and default output for `fim` command. Argument `-type depth|extent` is not required now. Argument `--with_domain` must be added if domain should be added as background in composite map.
- `fmt` argument for `fim` command now have `GTiff` option for GTiff format not `tif`. This is to be consistent with the GDAL raster drivers short names.
- flows2fim no longer need GDAL version 3.8.0 or greater when working with extent libraries. Any version above 3.4.0 should work.

### 0.3.0
*Compatible with outputs from Ripple1D Pipeline version unknown to 0.7.0*

- A new command `validate` has been added to validate FIM libraries (stored locally or on cloud) against rating curves table in a database. See Install.md for details on how to set it up
- Configuration through environment variables has been added
- Structured logging has been added. Following env variables control logging:
  - `F2F_LOG_LEVEL`: Set the logging level. Options are 'DEBUG', 'INFO', 'WARN', and 'ERROR'. Default is 'INFO'
  - `F2F_NO_COLOR`: Set to 'TRUE' to disable colored output. Default is 'FALSE'
- A bug has been fixed for `fim` command on extent libraries that was causing gaps in composite FIMs. This changes API for `fim` command. Argument `-type depth|extent` is required now.
- `-rel` argument from `fim` command is removed
- flows2fim need a version of GDAL 3.8.0 or greater when working with extent libraries


### 0.2.1
*Compatible with outputs from Ripple1D version unknown to 0.7.0*

- Fix bug that was assuming no data value -9999.0 for all raster types, the no data value is now inferred from FIM library rasters
- An error is raised in `fim` command when control file is empty

### 0.2.0
*Compatible with outputs from Ripple1D pipeline version unknown to 0.7.0*

- Default db table name changed from `conflation` to `network`
- Default `to_id` name changed from `conflation_to_id` to `updated_to_id`
- No more warnings for when stage difference is because of normal depth higher than target stage
- Add COG option for `fim` command

### 0.1.0