FROM ghcr.io/osgeo/gdal:ubuntu-small-3.8.5

# Install Go 1.24.1
RUN apt-get update && \
    apt-get install -y wget && \
    wget https://dl.google.com/go/go1.24.1.linux-$(dpkg --print-architecture).tar.gz -O /tmp/go.tar.gz && \
    tar -C /usr/local -xzf /tmp/go.tar.gz && \
    rm /tmp/go.tar.gz && \
    apt-get remove -y wget && \
    apt-get autoremove -y && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Set Go environment variables
ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH=/go

# Set working directory and copy Go mod files
WORKDIR /app
COPY go.mod go.sum ./

# Download and cache dependencies (assumes go.mod and go.sum are tidy)
RUN go mod download

# zip needed for creating release-assets
# git needed for version and tag information
RUN apt-get update && \
    apt-get install -y zip git && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# We need gdal_ls in path for validate function in cloud
RUN cp /usr/lib/python3/dist-packages/osgeo_utils/samples/gdal_ls.py /bin && \
    chmod +x /bin/gdal_ls.py

# Set git safe directory
RUN git config --global --add safe.directory /app
