### 0.2.1
*Compatible with outputs from Ripple1D version unknown to 0.7.0*

- Fix bug that was assuming no data value -9999.0 for all raster types, the no data value is now inferred from FIM library rasters
- An error is raised in `fim` command when control file is empty

### 0.2.0
*Compatible with outputs from Ripple1D version unknown to 0.7.0*

- Default db table name changed from `conflation` to `network`
- Default `to_id` name changed from `conflation_to_id` to `updated_to_id`
- No more warnings for when stage difference is because of normal depth higher than target stage
- Add COG option for `fim` command

### 0.1.0