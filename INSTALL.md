# Installation Instructions


## Docker
Coming soon

## Windows

1. **Install GDAL**
   - The easiest way is via [**OSGeo4W**](https://trac.osgeo.org/osgeo4w/#QuickStartforOSGeo4WUsers)

2. **Setup flows2fim**
    - Go to the [**Releases**](https://github.com/ar-siddiqui/flows2fim/releases) page and download the `flows2fim-windows-amd64.zip`
    - Unzip and copy `flows2fim.exe` into `C:\OSGeo4W\bin`

3. **(Optional) Enable `gdal_ls`**
   - This step is **only needed** if you plan to use the `flows2fim validate` command with a FIM library on cloud storage.

   _Your actual paths might be slightly different based on the version of python_
   - Copy `C:\OSGeo4W\apps\Python312\Lib\site-packages\osgeo_utils\gdal_ls.py` to `C:\OSGeo4W\apps\Python312\Scripts`.
   - In `C:\OSGeo4W\apps\Python312\Scripts`, make a copy of `gdal_merge.bat` → `gdal_ls.bat`.
   - Open `gdal_ls.bat` and replace all occurrences of `gdal_merge.py` with `gdal_ls.py`.

4. **Verify**
    - Open the **OSGeo4W Shell**.
    - Run `flows2fim --version` to confirm everything works.
    - Run `gdalinfo --version` to confirm everything works.
    - (Optional) Run `gdal_ls --version` if you set it up.


## Linux

1. **Install GDAL**

    _For Ubuntu_
   ```bash
   sudo apt-get update && sudo apt-get install -y gdal-bin
   ```

2. **Setup flows2fim**
    1. **Download**
        - Go to the [**Releases**](https://github.com/ar-siddiqui/flows2fim/releases) page and download the `flows2fim-linux-amd64.tar.gz`
        - Extract it and move to a directory in PATH (e.g., `/usr/local/bin`) and make it executable:
            ```bash
            tar -xvf flows2fim-linux-amd64.tar.gz
            sudo mv flows2fim /usr/local/bin/
            sudo chmod +x /usr/local/bin/flows2fim
            ```
    2. **Build from source**
        - cd into root of repository
        - Build container
            - `docker compose up -d`
        - Get flows2fim CONTAINER_ID
            - `CONTAINER_ID=$(docker ps -q | head -n 1)`
        - Issue build script
            - `docker exec $CONTAINER_ID /bin/bash -c "./scripts/build-linux-amd64.sh"` 
        - Shutdown container
            - `docker compose down`
        - Move executable to $PATH
            - `sudo mv builds/linux-amd64/flows2fim /usr/local/bin/`
            - `sudo chmod +x /usr/local/bin/flows2fim`
3. **(Optional) Enable `gdal_ls`**
   - Only required if using `flows2fim validate` with FIM libraries stored on cloud.
   - On Ubuntu/Debian systems:
     ```bash
     sudo cp /usr/lib/python3/dist-packages/osgeo_utils/samples/gdal_ls.py /usr/local/bin
     sudo chmod +x /usr/local/bin/gdal_ls.py
     ```

4. **Verify**
    - Run `flows2fim --version` to confirm everything works.
    - Run `gdalinfo --version` to confirm everything works.
    - (Optional) Run `gdal_ls.py --version` if you set it up.
