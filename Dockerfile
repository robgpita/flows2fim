FROM golang:1.22.1

WORKDIR /app
COPY go.mod go.sum ./

# Download and cache dependencies (assumes go.mod and go.sum are tidy)
RUN go mod download

# zip needed for creating release-assets
RUN apt-get update && \
    apt-get install zip -y && \
    apt-get install -y gdal-bin && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

RUN git config --global --add safe.directory /app

RUN cp /usr/lib/python3/dist-packages/osgeo_utils/samples/gdal_ls.py /bin && \
    chmod +x /bin/gdal_ls.py
