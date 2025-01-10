1. Download `flows2fim` executables from [Releases](https://github.com/ar-siddiqui/flows2fim/releases).

1. Install `GDAL` if you don't already have it. GDAL can be installed in a variety of ways.
    - On Windows: The easiest way is through `OSGeo4W` installer https://trac.osgeo.org/osgeo4w/#QuickStartforOSGeo4WUsers
    - On Ubuntu Linux: Run `apt-get update && apt-get install -y gdal-bin`

1. Make sure `flows2fim` and `GDAL` both are available in your Path.
   - On Windows: The easiest way is to place the downloaded `flows2fim.exe` file from step 1 in `C:\OSGeo4W\bin` and then use `OSGeo4W Shell` for executing `flows2fim`
   - On Linux: The simplest option is to place the downloaded `flows2fim` file in `/bin` folder

1. Add `gdal_ls` to Path (only required if using `validate` command with FIM library stored on the cloud).
    - On Windows:
         - Copy `C:\OSGeo4W\apps\Python312\Lib\site-packages\osgeo_utils\gdal_ls.py` to `C:\OSGeo4W\apps\Python312\Scripts` directory
         - Copy `C:\OSGeo4W\apps\Python312\Scripts\gdal_merge.bat` as `C:\OSGeo4W\apps\Python312\Scripts\gdal_ls.bat` and replace all occurences of `gdal_merge.py` with `gdal_ls.py` in the file
   - On Linux: Run `cp /usr/lib/python3/dist-packages/osgeo_utils/samples/gdal_ls.py /bin && chmod +x /bin/gdal_ls.py`
